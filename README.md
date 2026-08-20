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

## Continuous integration

CI runs on every push to `main` and on every pull request, via three GitHub
Actions workflows in [`.github/workflows/`](.github/workflows/). They are
independent and trigger on the same events, so they run **in parallel** as three
separate check runs:

| Workflow | File | Purpose |
| --- | --- | --- |
| **CI** | `ci.yml` | Build, lint, and test each service |
| **Security** | `security.yml` | Open-source security scanning across the repo |
| **CodeQL** | `codeql.yml` | GitHub semantic static analysis (Go, JS/TS) |

### How the pipeline is wired

GitHub Actions runs jobs in parallel unless `needs` links them; steps inside a
job always run in order. The overall shape:

```
On push / PR
│
├─ CI ───────────────────────────────────────────────────────
│    changes                    detect which service dirs changed
│       │ needs
│       ├──► go    ┐
│       ├──► rust  ├─ run in PARALLEL (only if their paths changed)
│       └──► js    ┘
│                    │ needs [go, rust, js]
│                    └──► ci-success   final gate — the required check
│
├─ Security ─────────────────────────────────────────────────
│    trivy-fs · hadolint · gitleaks · semgrep · govulncheck
│    cargo-audit · cargo-deny · pnpm-audit · dependency-review
│       └─ ALL PARALLEL (no ordering between them)
│
└─ CodeQL ───────────────────────────────────────────────────
     analyze [go, javascript-typescript]   → PARALLEL (matrix)
```

Two ordering points exist, both in the CI workflow. `changes` (using
[`dorny/paths-filter`](https://github.com/dorny/paths-filter)) runs first and
decides which language jobs are needed — touching only `js/` won't rebuild Go or
Rust. Then `ci-success` waits on all three and goes green only if none failed
(skipped jobs are allowed); mark **`ci-success`** as the single required status
check in branch protection.

Other efficiency measures: `concurrency` cancels superseded runs on the same
ref, dependency caches are keyed per language, and each workflow declares
least-privilege `permissions` (read-only by default, `security-events: write`
only where results are uploaded).

### CI workflow — build, lint, test

Each service has its own job. Every step must pass:

| Service | Format | Lint / static analysis | Test | Build |
| --- | --- | --- | --- | --- |
| **goapi** (Go) | `gofmt` | `go vet`, staticcheck, golangci-lint | `go test -race` + coverage | `go build ./...` |
| **rscounter** (Rust) | `cargo fmt --check` | `cargo clippy -D warnings` | `cargo test` | `cargo build --release --locked` |
| **js** (React/TS) | — | `oxlint` | `vitest run` | `tsc` typecheck + `pnpm build` |

### Security workflow — the tools

All scans use open-source tooling and run in parallel. Results that support
[SARIF](https://sarifweb.azurewebsites.net/) are uploaded to the repo's
**Security → Code scanning** tab.

| Tool | Scope | What it catches |
| --- | --- | --- |
| [**Trivy**](https://trivy.dev) | Whole repo (filesystem) | Vulnerable dependencies, leaked secrets, and misconfigurations (uploaded to SARIF). |
| [**Hadolint**](https://github.com/hadolint/hadolint) | The three Dockerfiles | Dockerfile best-practice and correctness issues. |
| [**Gitleaks**](https://github.com/gitleaks/gitleaks) | Full git history | Secrets (keys, tokens, passwords) committed at any point. |
| [**Semgrep**](https://semgrep.dev) | All source | Multi-language SAST using the public rule registry (`--config auto`). |
| [**govulncheck**](https://go.dev/blog/govulncheck) | goapi | Go code paths that actually reach a known-vulnerable symbol (Go vuln DB). |
| [**cargo-audit**](https://github.com/rustsec/rustsec) | rscounter | Rust dependencies with [RustSec](https://rustsec.org) advisories. |
| [**cargo-deny**](https://github.com/EmbarkStudios/cargo-deny) | rscounter | Advisories **plus** license policy and crate sources (config in `services/rscounter/deny.toml`). |
| **pnpm audit** | js | npm dependencies with known advisories (high+). |
| [**dependency-review**](https://github.com/actions/dependency-review-action) | Pull requests | Dependency changes that introduce vulnerable or badly-licensed packages. |

The Security workflow also runs **weekly on a schedule** so newly disclosed CVEs
are caught even without a code change.

#### Reported, not blocking — with one global report

Security findings often can't be fixed on the spot (an unpatched transitive
crate, a standard-library CVE fixed only in a newer toolchain). So every scanner
runs with `continue-on-error: true`: it **reports but does not fail the
pipeline**. Each job records its real outcome, and a final **`security-summary`**
job collects them into a single table on the run's **Job Summary** page — the one
place to see what's bad:

```
## 🔒 Security scan summary
| Tool        | Scope                          | Result                    |
| ----------- | ------------------------------ | ------------------------- |
| Trivy       | repo: deps, secrets, misconfig | ✅ Passed                 |
| govulncheck | Go stdlib + deps               | ❌ Findings — see job log |
| cargo-audit | Rust advisories                | ❌ Findings — see job log |
| ...         | ...                            | ...                       |
```

To make a scan **block** the pipeline instead, remove `continue-on-error: true`
from its job (or change the summary's final `exit 0` to `exit "$findings"` to
fail whenever any scanner reports something).

### CodeQL workflow

[CodeQL](https://codeql.github.com) is GitHub's semantic analysis engine: it
builds a queryable database of the code and runs data-flow queries to find
vulnerabilities in the project's *own* logic (injection, taint flows, etc.),
going deeper than pattern-based linters. It analyses **Go** and
**JavaScript/TypeScript**; Rust is covered instead by clippy, cargo-audit,
cargo-deny, and Trivy.


