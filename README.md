# goforge

An opinionated Go project scaffolder — like [go-blueprint](https://github.com/melkeydev/go-blueprint), but wired to one fixed stack and one set of conventions instead of a menu of choices.

Every project it generates gets: `cmd/api` + `cmd/ctl` entrypoints, `net/http` + a shared `httpx` kit, OIDC authentication, `viper` + `cobra` config, a working example domain (full CRUD vertical slice), `.golangci.yml`, a `Makefile`, GitHub Actions CI, and the `golang-*` conventions library (`.claude/skills`, `.agents/rules`) vendored in — so an AI agent (or a human) working in the generated repo already knows the conventions to follow.

## Install

Requires Go 1.26+.

```sh
go install github.com/VictorTarnovski/goforge/cmd/goforge@latest
```

This installs the `goforge` binary to `$(go env GOPATH)/bin` — make sure that's on your `PATH`.

Don't have Go installed? Grab it from [go.dev/dl](https://go.dev/dl/), or on Windows: `winget install GoLang.Go`.

## Usage

### Scaffold a new project

```sh
goforge new myapp --module github.com/you/myapp
```

Flags:

| Flag | Effect |
|---|---|
| `--module` | Go module path for the new project (**required**) |
| `--db` | Adds a Postgres-backed database layer: `pgx/v5` + `goose` migrations, `internal/db`, `internal/uow`, root-level `migrations/` |
| `--authz=openfga` | Adds OpenFGA permission checks (`internal/authz`) on top of the always-on OIDC authentication |
| `--deploy=nginx` | Adds a starter `deploy/nginx` reverse-proxy config |

`goforge new` also runs `go mod tidy` and best-effort auto-formats the result (`gofmt`, plus `golangci-lint run --fix` if that's on your `PATH`) — so a fresh scaffold is already buildable and lint-clean.

Without `--db`, the example domain's repository is an in-memory store (so the project still runs and is testable with zero setup); with `--db`, it's the real Postgres-backed version plus its migration.

### Generate a new domain

From the root of a project `goforge new` created:

```sh
goforge generate domain widget --fields name:string,count:int,active:bool,price:float,released:time
```

This generates a full vertical slice under `internal/widget/` — model, service, repository, HTTP handler, and a table-driven test — and **automatically wires it** into `cmd/api/main.go` (import, route registration, error classification). It adapts to how the project was scaffolded: with `--db` it also adds a numbered goose migration; without it, no migration is written. Re-running it for a domain that's already wired is a no-op.

Supported field types: `string`, `int`, `bool`, `float`, `time`.

## What gets scaffolded, layout-wise

```
cmd/api/            HTTP server entrypoint
cmd/ctl/             operational CLI (includes `migrate` when --db is set)
internal/<domain>/   one package per business domain (model + service + repository + HTTP handler + tests)
internal/httpx/      shared HTTP kit: JSON envelopes, RFC 7807 problem details, CORS
internal/authn/      OIDC bearer-token verification (always on)
internal/authz/      OpenFGA permission checks (only with --authz=openfga)
internal/db/         Postgres connection + goose migration runner (only with --db)
internal/uow/        unit-of-work over database/sql (only with --db)
migrations/          goose SQL migrations (only with --db)
build/               Dockerfile
deploy/nginx/        reverse-proxy config (only with --deploy=nginx)
docs/adr/            architecture decision records, with a starter template
.claude/skills/,
.agents/rules/        vendored Go conventions library (point-in-time snapshot, not auto-synced)
.golangci.yml, Makefile, .github/workflows/ci.yml, README.md, .env.example
```

`docker-compose.yml` is generated only when there's something to orchestrate (`--db` and/or `--authz=openfga`).

## Development

```sh
git clone https://github.com/VictorTarnovski/goforge
cd goforge
go build -o goforge.exe ./cmd/goforge
```

Templates live under `internal/templates/base` (the project skeleton) and `internal/templates/domain` (the generic vertical-slice generator, used both for the baked-in example domain and for `generate domain`). They're embedded into the binary via `go:embed`, so after editing a template you need to rebuild `goforge` for the change to take effect.
