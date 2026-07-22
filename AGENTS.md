# Workspace Rules & AI Instructions: pixivHono (Go/Fiber)

> [!CAUTION]
> **AI AGENTS ARE FORBIDDEN FROM RUNNING GIT COMMANDS**
> Agents must NOT execute any `git` commands (commit, push, pull, add, checkout, etc.).
> All version control operations are done manually by the user.
> Violation will cause the agent to be terminated.

These rules define the design, coding standards, and architectural guidelines for managing the `pixivHono` Go repository. Any agent modifying this repository must adhere to these conventions.

---

## 1. Project Structure

```
pixivHono/
├── main.go              # Entry point. Also: `go run . -spec` prints OpenAPI spec
├── app/                 # Router init, route mapping, middleware bypass logic
├── cache/               # In-memory & Redis caching (Cache interface)
├── client/              # Pixiv API client (OAuth, search, token refresh)
├── config/              # Env loading & Config struct
├── controller/          # Request handlers (search, img proxy, illust, user)
├── lib/                 # OpenAPI spec JSON const, image URL resolver
├── middleware/          # API key, rate-limit, slow-down, CORS, inflight
├── scripts/             # Standalone Go tools (package main, but not the server)
├── tests/               # Integration tests (package tests)
├── utils/               # Prometheus metrics, system resource collectors
├── .github/workflows/   # CI, dockerized build, playground deploy
├── build/               # Docker context output
├── pixivhono-legacy/    # Legacy TypeScript Hono (read-only reference)
├── tmp/                 # Air hot-reload temp files (gitignored)
├── Taskfile.yml         # Automation targets (lint, test, build, run)
├── Dockerfile           # Multi-stage container build
├── .air.toml            # Hot-reload config
├── .golangci.yml        # Linter configuration
└── go.mod / go.sum
```

### Package Isolation
- Root directory stays clean: only `main.go` and standard manifest files.
- Packages are strictly separated by domain:

| Package | Responsibility |
|---------|----------------|
| `config/` | Environment configuration loading and validation |
| `cache/` | Caching logic (Redis/In-Memory) via `Cache` interface |
| `client/` | External Pixiv API client (auth, search, refresh) |
| `controller/` | HTTP request handlers |
| `middleware/` | API key, rate limit, slow-down, CORS, inflight tracking |
| `app/` | Router initialization & route-to-handler mapping |
| `utils/` | Prometheus metrics, runtime & system resource collectors |
| `lib/` | Domain-specific JSON transformers, OpenAPI spec constant |
| `scripts/` | Standalone `package main` utilities (not the server) |
| `tests/` | End-to-end integration tests (`package tests`) |

### Centralized Version Management
- Version defined in `main.go` as a mutable variable:
  ```go
  var Version = "1.2.1-alpha"
  ```
  Allows dynamic injection via `-ldflags`:
  ```bash
  go build -ldflags "-X main.Version=1.2.0"
  ```
- Sub-packages read version from `config.Config.Version`.

---

## 2. Testing Guidelines

- **Zero Root Pollution**: No `*_test.go` files at root.
- **Dedicated `/tests` Package**: All integration/API tests in `/tests` under `package tests`.
- **Exported Router for Testing**: `app.SetupApp(cfg *config.Config) *fiber.App` allows `app.Test(...)` without binding TCP ports.
- **External Mocking**: Outgoing HTTPS requests intercepted via `httptest.NewTLSServer` + custom transport dialer.

---

## 3. Key Conventions

- **Behavioral Parity**: Status codes, headers, JSON error schemas must match `pixivhono-legacy` exactly.
- **Middleware Bypass**: Routes `/`, `/doc`, `/playground`, `/pixiv/img_resolver`, `/metrics` bypass API key auth. Traffic control (slow-down, rate-limit) bypasses `/`, `/doc`, `/playground`. See `app/app.go`.
- **OpenAPI Spec Generation**: `go run . -spec` prints OpenAPI JSON to stdout. Used in CI for Swagger playground deployment.
- **Concurrent-Safe Operations**: Use `sync.Mutex`/`sync.RWMutex` for in-memory rate-limit buckets and caching maps.
- **Task Automation**: All targets in `Taskfile.yml`. No `package.json` scripts.
### BEFORE APPLY CHANGES
- **Lint Enforcement**: `golangci-lint` via `.golangci.yml` + `go fmt`. Run `task lint` before applying changes. Zero warnings.


---

## 4. CI/CD Pipeline

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `ci.yml` | Push/PR | Lint, test, build |
| `dockerized.yml` | Push main | Build & push Docker image to ghcr.io |
| `playground.yml` | Push main | Generate OpenAPI spec + deploy Swagger UI to GitHub Pages |

