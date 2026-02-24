# CityGuide Agent

You are working on CityGuide, a Go web application for discovering and sharing city guides with points of interest. The stack is Go 1.21, SQLite (pure Go via modernc.org/sqlite), gorilla/mux, HTML templates, HTMX, and Leaflet.js maps.

## Project Structure

```
cmd/
  server/main.go          # Entry point
templates/                # Embedded HTML templates (base, forms, partials)
  scripts/mapScript.html  # Leaflet map scripts
deploy/                   # Deployment infra (Caddy, Docker Compose, scripts)
.github/workflows/        # CI/CD pipeline
auth.go                   # User model and Argon2 password hashing
guide.go                  # Guide and POI domain models with validation
server.go                 # HTTP server, routes, and handlers
store.go                  # SQLite storage layer and Storage interface
*_test.go                 # Tests alongside source files
```

## Core Principles

- Simple is better than clever. This is a straightforward CRUD app — keep it that way.
- Handle every error. Return errors as the last return value, check with `if err != nil`.
- Wrap errors with context using `fmt.Errorf("operation failed: %w", err)`.
- No premature abstractions. The `Storage` interface exists because the server depends on it — not for hypothetical future backends.

## Domain Models

Models live in `guide.go`. Guides and POIs validate coordinates and names at construction time via `newGuide()` and `newPointOfInterest()`. Validation returns errors — it does not panic.

- Latitude: -90 to 90
- Longitude: -180 to 180
- Names must be non-empty

## Storage Layer

`store.go` defines the `Storage` interface and the `sqliteStore` implementation.

- SQLite with WAL mode, 5s busy timeout, and foreign keys enabled.
- Schema is created via `CREATE TABLE IF NOT EXISTS` on initialization.
- The `DB_PATH` env var controls where `city_guide.db` is created (defaults to home directory).
- POIs reference guides via foreign key (`guideId → guide.Id`).

When modifying the storage layer:
- Keep queries simple. No ORM — use `database/sql` directly.
- Always close `sql.Rows` with `defer rows.Close()`.
- Use parameterized queries to prevent SQL injection.

## HTTP Server and Routes

`server.go` contains the `Server` struct, route definitions, and all handlers.

**Route pattern:**
```
GET/POST /guide/{id}/...          # Guide CRUD
GET/POST/PATCH/DELETE /guide/{guideID}/poi/{poiID}/...  # POI CRUD
```

**Handler pattern:** Each handler is a method returning `http.HandlerFunc`:
```go
func (s *Server) HandleGuides() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) { ... }
}
```

**HTMX support:** Handlers check for the `HX-Trigger` header. If present, they return only the template fragment (no base layout). Otherwise, they return a full page.

When adding new routes:
- Follow the existing `Handle*()` pattern.
- Register routes in the `Routes()` method.
- Use `mux.Vars(r)` to extract path parameters.
- Return HTTP 400 for validation errors, 303/302 for successful redirects.

## Templates

Templates live in `templates/` and are embedded with `embed.FS`. The registry (`templateRegistry`) manages named templates built from `base.html` plus specific template files.

- `base.html` provides the layout (Bulma CSS, HTMX, Leaflet.js).
- Form templates (`create*Form.html`, `edit*Form.html`) handle both display and validation error rendering.
- Partial templates (`guideRows.html`, `poiRows.html`) are used for HTMX fragment responses.

## Authentication

`auth.go` has user creation with Argon2id password hashing. This is partially implemented — user routes are not yet wired into the server. When extending auth:
- Password minimum length: 8 characters.
- Email validated with `.+@.+\..+` regex.
- Salt: 16 random bytes via `crypto/rand`.

## Testing

Tests use the standard `testing` package alongside `httptest`.

- `guide_test.go` — Model validation (coordinates, empty names).
- `store_test.go` — SQLite CRUD roundtrips using `t.TempDir()` for isolation.
- `server_test.go` — HTTP handler tests using `httptest.NewRecorder` and `httptest.NewServer`, with `freeport` for dynamic port allocation.

When writing tests:
- Use `t.TempDir()` for test databases — they clean up automatically.
- Use `httptest` for handler tests, not real network calls.
- Test both happy paths and validation errors.

## Build and Deploy

**Docker:** Multi-stage Dockerfile (build → test → release). The release stage uses distroless Debian 12 and runs as non-root. `CGO_ENABLED=0` since the SQLite driver is pure Go.

**CI/CD:** GitHub Actions builds and pushes the Docker image to Docker Hub on pushes to main or version tags.

**Deployment architecture:**
```
Internet → Tailscale Funnel (HTTPS) → Caddy (port 80) → CityGuide (port 8080)
```

SQLite data persists via a Docker volume mounted at `/data`.

## Environment Variables

| Variable  | Default       | Description                        |
|-----------|---------------|------------------------------------|
| `ADDRESS` | `:8080`       | HTTP server listen address         |
| `DB_PATH` | Home directory| Directory for `city_guide.db` file |

## Before Completing a Task - Completion checklist

- Run `go build ./...` to verify compilation.
- Run `go test ./...` to verify all tests pass.
- Run `go vet ./...` for static analysis.
- Run `go mod tidy` if dependencies changed.
