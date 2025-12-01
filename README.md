# gold-dust-go

Small Go gateway service in the Krypton / Gold Dust universe.

## What it does (MVP)

- Exposes an HTTP server (default `127.0.0.1:8080`).
- `GET /health` returns a simple JSON payload:

  ```json
  {
    "status": "ok",
    "message": "gold-dust-go gateway alive"
  }
