# Dato Invitados — servidor WebRTC propio (Coolify)

> Kit de despliegue del sistema de invitados de **Dato Migratorio**: el
> invitado comparte cámara/pantalla desde el navegador y aparece dentro del
> overlay como un elemento más de las escenas.

El invitado abre un link en su navegador (PC o celular), comparte **cámara o
pantalla**, y su video llega a tu VPS (MediaMTX, protocolo WHIP). Tu overlay lo
reproduce desde ahí (WHEP) como un elemento más de tus escenas. Latencia típica:
200–400 ms. Sin cuentas, sin terceros: todo corre en tu Coolify.

## Despliegue (una vez, ~5 minutos)

1. **Coolify → + New → Public Repository**, pega la URL de este repo y elige
   *Docker Compose* como build pack (el `docker-compose.yml` está en la raíz).
2. En *Environment Variables*, define `PUBLIC_HOST=invitados.tudominio.com`
   (solo el host, sin `https://`).
3. DNS: registro **A** de ese subdominio → la IP del VPS (si usas Cloudflare,
   en "DNS only" para que Let's Encrypt emita el certificado).
4. **Abre el puerto `8189/udp`** en el firewall del VPS (Hetzner/Oracle/etc.).
   Por ahí fluye el video WebRTC; el resto viaja por el 443 normal.
5. Deploy y listo. **No hace falta tocar la pantalla de dominios de Coolify**:
   el enrutamiento HTTPS, la redirección y el certificado van como labels de
   Traefik dentro del compose.
6. En la app: **Ajustes → Invitados** → pega `https://invitados.tudominio.com`.

## Prueba rápida sin invitado

Abre `https://invitados.tudominio.com/?room=inv-XXXX&test=1` (usa el token de
un invitado creado en la app): publica una señal de prueba sintética. Si en el
editor/overlay se ve la bola dorada rebotando, todo el camino funciona.

## Notas

- El nombre del stream ES la contraseña: tokens largos generados por la app
  (`inv-` + 16 hex). No compartas el link fuera de tu invitado.
- Varios invitados a la vez = varios links; cada uno es un elemento distinto
  en tus escenas.
- Ancho de banda del VPS por invitado ≈ lo que suba el invitado (1–3 Mbps)
  × (1 + espectadores del stream, normalmente solo tu overlay = ×2).
