# Workspace Rules & AI Instructions: pixivHono (Go/Fiber)

These rules define the design, coding standards, and architectural guidelines for managing the `pixivHono` Go repository. Any agent modifying this repository must adhere to these conventions.

---

## 1. Architectural Standards

### Directory Structure & Package Isolation
- The root directory must remain clean of loose utility and configuration files. It should only contain `main.go` and standard project manifest files.
- Modulize the application strictly into distinct packages:
  - `config/`: Environment configuration loading and validation.
  - `cache/`: Caching logic (Redis/In-Memory).
  - `client/`: External client APIs (Pixiv integration).
  - `controller/`: Request handlers.
  - `middleware/`: Custom traffic and auth filters (API key, rate limit, CORS).
  - `app/`: Router initialization and route mapping to prevent package `main` import locks.
  - `utils/`: System resource scraping and metrics collectors.
  - `lib/`: Domain-specific JSON transformers (image URL resolver).

### Centralized Version Management
- The application version must be defined at the root of the workspace inside `main.go` as a mutable package-level variable:
  ```go
  var Version = "1.1.0-alpha"
  ```
  *Reason*: This permits dynamic compile-time version injection via standard compiler flags during the build process:
  ```bash
  go build -ldflags "-X main.Version=1.2.0"
  ```
- Any sub-package requiring the version string must read it dynamically via the `config.Config` configuration object.

---

## 2. Testing Guidelines

- **Zero Root Pollution**: No test files (`*_test.go`) are permitted at the root directory of the workspace.
- **Dedicated `/tests` Package**: All end-to-end integration and API tests must reside in a dedicated `/tests` directory under `package tests`.
- **Exposed Router for Testing**: The application routing setup must be exported from the `app` package via `SetupApp(cfg *config.Config) *fiber.App`. This allows external test runners to spin up mock client requests context-safely using `app.Test(...)` without binding actual TCP ports.
- **External Mocking**: Outgoing HTTPS requests (e.g. auth and image servers) must be intercepted using a local `httptest.NewTLSServer` and a custom transport dialer redirect rather than introducing network dependencies in tests.

---

## 3. Parity & Coding Conventions

- **1:1 Behavioral Parity**: Behavior, status codes, headers, and JSON error schemas must match the legacy TypeScript Hono implementation (`pixivhono-legacy`) exactly.
- **Concurrent-Safe Operations**: Use locks (`sync.Mutex` or `sync.RWMutex`) to safeguard in-memory data structures like sliding rate-limit buckets and caching maps against concurrent race conditions.
- **Task Automation**: Use `Taskfile.yml` to define automation targets. Do not fall back to `package.json` scripts.
- **Hot Reloader**: Exclude `pixivhono-legacy` (which contains legacy `node_modules`), `/tests`, `/build`, and `tmp` directories inside `.air.toml` to prevent infinite watch loops and high CPU usage.
- **Lint Enforcement**: Code style and static analysis are enforced via `golangci-lint` (using `.golangci.yml` guidelines) and `go fmt`.

