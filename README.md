# vc-sampleapp-go-rust-react

A tiny full-stack sample that counts daily visits. Each time the page loads it
records a visit and shows today's date, today's count, the all-time total, and a
per-day history.

## Architecture

```
  React (js)  ──►  goapi (Go)  ──►  rscounter (Rust)  ──►  Postgres
    :3000            :8080              :8081                :5432
   the UI         HTTP API for       records events,      stores one row
                  the frontend       counts per day       per visit
```

- **js** — React + Rsbuild (Rspack) frontend. Calls goapi, proxying `/api`.
- **goapi** — Go stdlib HTTP API. The only service the frontend talks to; it
  forwards to rscounter.
- **rscounter** — Rust (axum + sqlx). Owns the database; records visits and
  computes counts.
- **Postgres** — one `events` row per visit; counts are `COUNT(*)` queries.

### Endpoints

goapi (called by the frontend):

- `GET /api/today` — record a visit, return `{ date, day_count, total_count }`
- `GET /api/history` — per-day totals, newest first: `[{ date, count }]`

rscounter (internal): `POST /events`, `GET /counter?date=`, `GET /history`.

## Getting started

Requires [Podman](https://podman.io). Then, from the repo root:

```sh
make up      # build images, start Postgres + all three services
             # open http://localhost:3000
make down    # stop and remove everything (keeps the data volume)
```

`make help` lists every target.

### Running a service locally

Each service also runs on the host — handy for fast iteration. Start the
database first, then the services it depends on:

```sh
make db                                  # Postgres in a container
cd services/rscounter && cargo run       # :8081
cd services/goapi     && go run .        # :8080
cd js                 && pnpm dev        # :3000 (Rspack dev server, HMR)
```

Local and containerized services interoperate, so you can mix them (e.g. run the
frontend with `pnpm dev` against the rest in containers).
