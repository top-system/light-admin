# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Light Admin Server is a Go backend management system scaffold with RBAC permission management, built on Echo + GORM + Casbin + Uber-FX. Module path: `github.com/top-system/light-admin`. Requires Go 1.24+.

## Common Commands

```bash
make start              # Run the server (go run ./main.go runserver ...)
make build              # Compile binary to ./build/echo-admin
make migrate            # Run database migrations
make setup              # Initialize menus and admin user from config
make swagger            # Generate Swagger docs (swag init)
make fmt                # Format Go code

# Run all tests
go test ./...

# Run a specific test
go test ./tests -run TestQueueBasic -v

# Run benchmarks
go test ./tests -bench=BenchmarkQueueThroughput
```

## Architecture

### Dependency Injection (Uber-FX)

Everything is wired through Uber-FX modules. The composition chain is:

```
main.go → cmd (Cobra CLI) → bootstrap.Module
  └─ lib.Module        (Config, Logger, Database, Cache, Captcha, WebSocket, Extras)
  └─ api.Module        (Middlewares, System module, Platform module, Routes)
  └─ fx.Invoke(bootstrap)  (lifecycle hooks: start server, setup middlewares/routes)
```

When adding a new provider or service, register it in the appropriate FX module (`lib/lib.go`, `api/module.go`, or the domain-specific `module.go`).

### Layered Architecture per Domain Module

Each domain module (`api/system/`, `api/platform/`) follows a strict 4-layer pattern, each with its own FX module:

```
route/ → controller/ → service/ → repository/
```

- **route**: Defines Echo endpoints, implements `Route` interface with `Setup()` method
- **controller**: Handles HTTP request/response, validation, calls services
- **service**: Business logic
- **repository**: GORM database queries

To add a new API resource: create files in all 4 layers, register each in its layer's FX module, then add the route to the module's `Routes` slice.

### Module System

Two domain modules exist under `api/`:
- **system** — Core RBAC: users, roles, menus, departments, dicts, configs, logs, notices, auth
- **platform** — Extended features: file upload, WebSocket, task queue management, downloads

To add a new domain module: create it under `api/`, register it in `api/module.go`, and add its routes to the `Routes` struct.

### Middleware Chain

Middlewares execute in this order (configured in `api/middlewares/`):
1. **Core** — Framework defaults (recover, request ID)
2. **Zap** — Structured request logging
3. **CORS** — Cross-origin headers
4. **Auth** — JWT token validation (skips paths in `config.Auth.IgnorePathPrefixes`)
5. **Casbin** — RBAC policy enforcement (skips paths in `config.Casbin.IgnorePathPrefixes`)
6. **Log** — Operation audit logging to database

### Database

GORM supports MySQL, PostgreSQL, and SQLite (configured via `config.Database.Engine`). Database initialization logic with driver-specific setup is in `lib/db.go`. Table prefix is configurable (`config.Database.TablePrefix`). Models live in `models/system/` and `models/platform/`. Base model types (DateTime, UUID, JSONB) are in `models/database/`.

### Configuration

Viper-based YAML config at `config/config.yaml` (copy from `config.yaml.default`). Key sections: `Http`, `Log`, `SuperAdmin`, `Auth`, `Captcha`, `Casbin`, `Cache` (memory or redis), `Database`, `OSS` (local/minio/aliyun). Extended features (`Queue`, `Crontab`, `Downloader`) are in `config/extras.yaml`.

### Extended Features (lib/extras.go)

Optional features loaded via `lib.ExtrasModule`:
- **Task Queue** (`pkg/queue/`) — Worker pool with retry, persistent via GORM
- **Crontab** (`pkg/crontab/`) — Scheduled tasks via robfig/cron
- **Downloader** (`pkg/downloader/`) — aria2/qBittorrent RPC integration

Each is enabled/disabled independently via config.

### WebSocket

WebSocket connections at `/ws` bypass the Echo middleware stack entirely — they're intercepted at the `http.Server` level in `bootstrap/bootstrap.go` and handled directly by `WebSocketController`. Uses STOMP protocol.

### Cache

Two backends (`lib/cache.go`): in-memory (default) or Redis. Casbin permissions are cached for performance; cache is invalidated on policy changes.

### CLI Commands (cmd/)

Cobra subcommands: `runserver` (start API server), `migrate` (auto-migrate all models), `setup` (seed menus from `config/menu.yaml` and create super admin).

## API Routes

All API endpoints are prefixed with `/api/v1/`. Route definitions are in `api/system/route/` and `api/platform/route/`, one file per resource.
