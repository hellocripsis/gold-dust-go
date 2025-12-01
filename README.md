# gold-dust-go

[![CI](https://github.com/hellocripsis/gold-dust-go/actions/workflows/ci.yml/badge.svg)](https://github.com/hellocripsis/gold-dust-go/actions/workflows/ci.yml)

`gold-dust-go` is a small Go HTTP gateway that exposes:

* `GET /health` – returns a gateway status and a Krypton entropy snapshot.
* `POST /jobs` – accepts a job request and returns an **accepted / throttled / denied** decision.

It is designed to sit in front of the Rust project [`krypton-entropy-core`](https://github.com/hellocripsis/krypton-entropy-core) and the Python boundary layer [`krypton-boundary-orchestrator`](https://github.com/hellocripsis/krypton-boundary-orchestrator).

The gateway is **portfolio-safe** and uses only OS RNG–based entropy via Krypton’s public binary.

---

## Architecture

Current wiring:

* **Rust**: `krypton-entropy-core`

  * Binary: `entropy_health` (JSON).
  * Emits entropy metrics + `Keep / Throttle / Kill` decision.

* **Go**: `gold-dust-go`

  * HTTP server:

    * `GET /health` – includes a `krypton` field with the latest entropy snapshot.
    * `POST /jobs` – runs a simple job decision backed by Krypton health.

* **Python**: `krypton-boundary-orchestrator`

  * Talks to this gateway via `/health` and `/jobs`.
  * Uses decisions to gate jobs and run telemetry loops.

Data flow:

```text
entropy_health (Rust) ──▶ gold-dust-go (Go HTTP /health, /jobs) ──▶ Python orchestrator
```

---

## Configuration

The gateway is configured via environment variables:

* `GOLD_DUST_KRYPTON_MODE`:

  * `none`   – no external Krypton calls, use stub health.
  * `http`   – call a remote HTTP endpoint (e.g. another gateway).
  * `binary` – execute the `entropy_health` binary directly.

* `GOLD_DUST_KRYPTON_BIN` (when `mode=binary`):

  * Absolute path to the `entropy_health` binary, for example:

    ```text
    /home/youruser/dev/krypton-entropy-core/target/debug/entropy_health
    ```

* `GOLD_DUST_KRYPTON_URL` (when `mode=http`):

  * Base URL for an HTTP health endpoint, e.g. `http://127.0.0.1:3000/health`.

Internally:

* `internal/config` – loads configuration (env + defaults).
* `internal/krypton` – client for fetching entropy health.
* `internal/jobs` – maps Krypton decisions to job decisions.

---

## Quickstart

### 1. Clone and build

```bash
cd ~/dev
git clone git@github.com:hellocripsis/gold-dust-go.git
cd gold-dust-go
go mod tidy
```

### 2. Start with Krypton binary mode

Assuming you already have `krypton-entropy-core` built and `entropy_health` available:

```bash
cd ~/dev/gold-dust-go
GOLD_DUST_KRYPTON_MODE=binary \
GOLD_DUST_KRYPTON_BIN=/home/youruser/dev/krypton-entropy-core/target/debug/entropy_health \
go run ./cmd/gateway
```

By default the gateway listens on `127.0.0.1:8080`.

### 3. Check `/health`

In another terminal:

```bash
curl http://127.0.0.1:8080/health
```

Example response (shape only):

```json
{
  "status": "ok",
  "message": "gold-dust-go gateway alive",
  "krypton": {
    "samples": 2048,
    "mean": 0.5008,
    "variance": 0.0038,
    "jitter": 0.0496,
    "decision": "Keep",
    "source": "binary:/home/youruser/dev/krypton-entropy-core/target/debug/entropy_health",
    "at": "2025-12-01T19:37:12.468988532Z"
  },
  "addr": "127.0.0.1:8080",
  "krypton_mode": "binary"
}
```

### 4. Submit a job via `/jobs`

```bash
curl -s -X POST http://127.0.0.1:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"job_id":"demo-job","payload":{"foo":"bar"}}'
```

Example response:

```json
{
  "job_id": "demo-job",
  "decision": "accepted",
  "krypton": {
    "samples": 2048,
    "mean": 0.5029,
    "variance": 0.0038,
    "jitter": 0.0496,
    "decision": "Keep",
    "source": "binary:/home/youruser/dev/krypton-entropy-core/target/debug/entropy_health",
    "at": "2025-12-01T19:51:49.820535855Z"
  }
}
```

Internally, the job decision is derived from the Krypton `decision`:

* `Keep`     → `"accepted"`
* `Throttle` → `"throttled"`
* `Kill`     → `"denied"`

---

## Development

Format and test:

```bash
cd ~/dev/gold-dust-go
go fmt ./...
go test ./...
```

(At the moment tests are minimal; this repo is primarily a wiring example.)

Typical loop when iterating:

1. Start the gateway with your desired Krypton mode.
2. Hit `/health` to verify wiring.
3. Hit `/jobs` to exercise decision logic.
4. Optionally point `krypton-boundary-orchestrator` at this gateway for full Rust → Go → Python integration.

---

## What this demonstrates (for reviewers)

This repo is structured as a small but realistic gateway service. It shows that the author:

* Writes idiomatic Go services with `cmd/` + `internal/` layout.
* Integrates with an existing Rust entropy core via binary or HTTP.
* Exposes clean HTTP APIs (`/health`, `/jobs`) with JSON contracts.
* Uses configuration via environment variables instead of hardcoding paths.
* Treats even small components as production-style services (CI, formatting, wiring to other languages).
