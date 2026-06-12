// chatrelay — puente mínimo de mensajes por sala para la cabina del invitado.
// La app del operador se conecta como "pub" (con el token de la sala, que ya
// es secreto) y empuja el feed de chat; las cabinas se conectan como "sub" y
// lo reciben. Lo que un sub envía (sugerencias) llega a los pubs. Sin estado,
// sin persistencia: si nadie escucha, los mensajes se descartan.
package main

import (
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/gorilla/websocket"
)

var roomRe = regexp.MustCompile(`^inv-[a-z0-9]{12,}$`)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true }, // token = auth
}

type client struct {
	conn *websocket.Conn
	out  chan []byte
	pub  bool
}

type room struct {
	mu      sync.Mutex
	clients map[*client]bool
}

var (
	mu    sync.Mutex
	rooms = map[string]*room{}
)

func getRoom(name string) *room {
	mu.Lock()
	defer mu.Unlock()
	r := rooms[name]
	if r == nil {
		r = &room{clients: map[*client]bool{}}
		rooms[name] = r
	}
	return r
}

func (r *room) join(c *client)  { r.mu.Lock(); r.clients[c] = true; r.mu.Unlock() }
func (r *room) leave(c *client) { r.mu.Lock(); delete(r.clients, c); r.mu.Unlock() }

// fan envía msg a los clientes del otro lado (pub→subs, sub→pubs).
func (r *room) fan(from *client, msg []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for c := range r.clients {
		if c.pub == from.pub {
			continue
		}
		select {
		case c.out <- msg:
		default: // lento → que pierda mensajes antes que frenar la sala
		}
	}
}

func serveWS(w http.ResponseWriter, req *http.Request) {
	name := req.PathValue("room")
	if !roomRe.MatchString(name) {
		http.Error(w, "sala inválida", http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	c := &client{conn: conn, out: make(chan []byte, 256), pub: req.URL.Query().Get("role") == "pub"}
	r := getRoom(name)
	r.join(c)
	defer func() { r.leave(c); conn.Close() }()

	go func() {
		for msg := range c.out {
			if conn.WriteMessage(websocket.TextMessage, msg) != nil {
				return
			}
		}
	}()
	conn.SetReadLimit(64 << 10)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			close(c.out)
			return
		}
		r.fan(c, msg)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws/{room}", serveWS)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	log.Println("chatrelay en :8085")
	log.Fatal(http.ListenAndServe(":8085", mux))
}
