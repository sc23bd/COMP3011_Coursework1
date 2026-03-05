# COMP3011 Coursework 1 — RESTful API (Gin / Go + PostgreSQL)

A production-ready RESTful API in Go using the
[Gin](https://github.com/gin-gonic/gin) framework backed by **PostgreSQL**
(via [`lib/pq`](https://github.com/lib/pq)).  Every architectural decision is
explicitly mapped to one or more of the
**Six Guiding Principles of REST** (Fielding, 2000).

The server falls back automatically to an in-memory store when no
`DATABASE_URL` is configured, so the full test suite runs without a database.

---

## Six Guiding Principles of REST

| # | Principle | How this template addresses it |
|---|-----------|-------------------------------|
| 1 | **Client–Server** | HTTP handlers (`internal/handlers`) are completely decoupled from any client implementation. The server exposes a uniform HTTP interface; clients are free to be web browsers, mobile apps, or CLI tools. |
| 2 | **Stateless** | No server-side session state exists. Every request must carry all information needed to be processed (validated body / path parameters). Authentication is token-based using JWT — all user identity is carried in the self-describing token, not in server-side sessions. |
| 3 | **Cacheable** | The `CacheControl` middleware sets `Cache-Control: public, max-age=60` on `GET`/`HEAD` responses, enabling clients and intermediary caches to store them. Mutating methods (`POST`, `PUT`, `DELETE`) are marked `no-store`. |
| 4 | **Uniform Interface** | Resources are identified by versioned URIs (`/api/v1/items/{id}`). Standard HTTP verbs (`GET`, `POST`, `PUT`, `DELETE`) map to CRUD operations. Response bodies include HATEOAS hypermedia links so clients can discover related actions without out-of-band knowledge. |
| 5 | **Layered System** | Middleware (`RequestID`, `Logger`, `JWTAuth`, `CacheControl`, `Recovery`) forms transparent processing layers between the network and the handlers. The same binary runs correctly behind a load-balancer or reverse proxy. |
| 6 | **Code on Demand** *(optional)* | Not implemented by default. The architecture supports it — a handler could return executable JavaScript or WebAssembly to extend client functionality when required. |

---

## Project Layout

```
.
├── cmd/
│   └── server/
│       └── main.go              # Entry point — reads PORT, JWT_SECRET, DATABASE_URL env vars
├── internal/
│   ├── auth/
│   │   └── jwt.go               # JWT token generation and validation
│   ├── db/
│   │   ├── db.go                # PostgreSQL connection helper (Connect / ConnectFromEnv)
│   │   ├── item_repo.go         # PostgreSQL ItemRepo — implements ItemRepository
│   │   └── user_repo.go         # PostgreSQL UserRepo — implements UserRepository
│   ├── handlers/
│   │   ├── auth.go              # Authentication endpoints (register, login)
│   │   ├── handlers.go          # Item CRUD handlers + repository interfaces + in-memory Store
│   │   └── handlers_test.go     # Handler tests (run against the in-memory Store)
│   ├── middleware/
│   │   ├── auth.go              # JWT authentication middleware
│   │   └── middleware.go        # RequestID, Logger, CacheControl
│   ├── models/
│   │   ├── errors.go            # Shared sentinel errors (ErrNotFound, ErrConflict)
│   │   ├── item.go              # Item domain model + request/response types
│   │   └── user.go              # User domain model + auth request/response types
│   └── router/
│       └── router.go            # Wires middleware, repositories, and routes together
├── migrations/
│   └── 001_initial_schema.sql   # Idempotent DDL — users and items tables + indexes
├── docker-compose.yml           # PostgreSQL 16 for local development / integration testing
├── go.mod
├── go.sum
└── README.md
```

---

## Getting Started

### Prerequisites

* Go 1.24 or later
* **PostgreSQL 14+** *(optional — the server uses an in-memory store when
  `DATABASE_URL` is not set)*
* **Docker & Docker Compose** *(optional — for the quickest local database
  setup)*

### Option A — run without a database (in-memory store)

The server automatically falls back to an in-memory store when `DATABASE_URL`
is absent.  This is the default for local development and all unit tests.

```bash
# Requires only JWT_SECRET; no database needed
JWT_SECRET=your-secret-key go run ./cmd/server

# Custom port
JWT_SECRET=your-secret-key PORT=9090 go run ./cmd/server
```

### Option B — run with Docker Compose (recommended)

Create a `.env` file in the project root with your configuration:

```bash
# .env
JWT_SECRET=your-secret-key
db_user=<username for db>
db_password=<password for db>
db_name=<name for db>
db_url="postgres://${db_user}:${db_password}@db:5432/${db_name}?sslmode=disable"
```

**Run both the application and database:**

```bash
# Start both containers (schema is applied automatically on first start)
docker compose up --build -d

# View logs
docker compose logs -f

# Stop everything
docker compose down

# Stop and remove all data
docker compose down -v
```

**Run only the database container:**

```bash
# Start only PostgreSQL
docker compose up db -d

# Run the application locally, connecting to the containerized database
DATABASE_URL="postgres://<db_user>:<db_password>@localhost:5432/<db_name>?sslmode=disable" \
  JWT_SECRET=<JWT_SECRET> \
  go run ./cmd/server
```

**Run only the application container:**

```bash
# Start only the app (requires a PostgreSQL instance already running elsewhere)
# Update db_url in .env file to point to your external database
docker compose up app -d
```

### Option C — run with an existing PostgreSQL instance

```bash
# 1. Apply the schema (safe to run multiple times — all statements are IF NOT EXISTS)
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql

# 2. Start the server
DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" \
  JWT_SECRET=your-secret-key \
  go run ./cmd/server
```

### Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | **Yes** (production) | — | Secret key used to sign JWT tokens. Set `DEV_MODE=true` to auto-generate a random one for development (not safe for production). |
| `PORT` | No | `8080` | TCP port the server listens on |
| `DATABASE_URL` | No | *(in-memory)* | libpq connection string for PostgreSQL |
| `DEV_MODE` | No | — | Set to `true` to auto-generate `JWT_SECRET` in development |

### Run the tests

```bash
go test ./...
```

Tests run against the in-memory store; no database connection is required.

### Build a binary

```bash
go build -o api-server ./cmd/server
./api-server
```

---

## Database

### Schema

The schema lives in `migrations/001_initial_schema.sql`.  Every statement uses
`IF NOT EXISTS`, so the file is safe to re-apply.

```sql
-- users: bcrypt-hashed passwords only — plain text never stored
CREATE TABLE IF NOT EXISTS users (
    username      VARCHAR(50)  PRIMARY KEY,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- items: auto-incrementing SERIAL primary key
CREATE TABLE IF NOT EXISTS items (
    id          SERIAL       PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

Indexes are created on `items.updated_at DESC` (used for `Last-Modified`
header ordering) and `items.name` (future filtering / search support).

### Connection pooling

`internal/db/db.go` configures the `*sql.DB` pool:

| Setting | Value | Rationale |
|---------|-------|-----------|
| `MaxOpenConns` | 25 | Hard cap on concurrent DB connections |
| `MaxIdleConns` | 5 | Keep a small warm pool to reduce connection-setup latency |
| `ConnMaxLifetime` | 5 min | Recycle connections before load-balancer / firewall idle limits are hit |

### Repository pattern

Two interfaces live in `internal/handlers/handlers.go`:

```go
type ItemRepository interface { ... }
type UserRepository interface { ... }
```

`router.New` receives a `*sql.DB`; when it is non-nil the PostgreSQL
implementations (`db.ItemRepo`, `db.UserRepo`) are wired in.  When it is nil
the in-memory `handlers.Store` is used instead — no handler code changes.

---

## API Reference

Base URL: `http://localhost:8080/api/v1`

### Authentication Resource

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/register` | Register a new user account |
| `POST` | `/auth/login` | Login and receive JWT token |

### Items Resource

| Method | Path | Description | Auth Required |
|--------|------|-------------|---------------|
| `GET` | `/items` | List all items | No |
| `POST` | `/items` | Create a new item | **Yes** |
| `GET` | `/items/:id` | Get a single item | No |
| `PUT` | `/items/:id` | Replace an existing item | **Yes** |
| `DELETE` | `/items/:id` | Delete an item | **Yes** |

### Request / Response Examples

**Register a user**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"secret123"}'
```

```json
{
  "message": "user created successfully",
  "username": "john",
  "links": [
    {"rel":"login","href":"/api/v1/auth/login","method":"POST"}
  ]
}
```

**Login**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"secret123"}'
```

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "links": [
    {"rel":"items","href":"/api/v1/items","method":"GET"}
  ]
}
```

**List items** (no authentication required)

```bash
curl http://localhost:8080/api/v1/items
```

```json
{
  "data": [ ... ],
  "links": [
    {"rel":"self",   "href":"/api/v1/items","method":"GET"},
    {"rel":"create", "href":"/api/v1/items","method":"POST"}
  ]
}
```

**Create an item** (requires JWT token)

```bash
curl -X POST http://localhost:8080/api/v1/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"name":"Widget","description":"A sample widget"}'
```

```json
{
  "id": "1",
  "name": "Widget",
  "description": "A sample widget",
  "createdAt": "2024-01-01T12:00:00Z",
  "updatedAt": "2024-01-01T12:00:00Z",
  "links": [
    {"rel":"self",   "href":"/api/v1/items/1","method":"GET"},
    {"rel":"update", "href":"/api/v1/items/1","method":"PUT"},
    {"rel":"delete", "href":"/api/v1/items/1","method":"DELETE"}
  ]
}
```

### Response Headers

| Header | Description |
|--------|-------------|
| `X-Request-ID` | Unique ID for each request (traceability) |
| `Authorization` | Bearer token required for mutation operations (POST, PUT, DELETE) |
| `Cache-Control` | `public, max-age=60` on GET; `no-store` on mutations |
| `Location` | Set to the new resource URI on `201 Created` |

---

## Extending the Project

1. **Add a new resource** — create a handler file in `internal/handlers/`,
   define a new repository interface, add routes in `internal/router/router.go`.
2. **Add a database migration** — create the next numbered SQL file in
   `migrations/` (e.g. `002_add_tags.sql`) and apply it with `psql`.
3. **Add role-based access control** — extend the JWT claims in
   `internal/auth/jwt.go` to include roles, then add middleware to check
   permissions before reaching handlers.
4. **Add pagination** — the `items_updated_at_idx` and `items_name_idx`
   indexes already support efficient `LIMIT`/`OFFSET` or cursor-based queries;
   extend `ListItems` in `internal/db/item_repo.go` and the route handler.
5. **Add a caching layer** — introduce a `CachedItemRepo` that wraps
   `ItemRepository`; swap it in `internal/router/router.go` without touching
   any handler code.
6. **Configuration** — read additional settings from environment variables or
   a config file in `cmd/server/main.go`.

---

## International Football Results Feature

This feature exposes data from the
[International football results from 1872 to 2025](https://www.kaggle.com/datasets/martj42/international-football-results-from-1872-to-2017)
Kaggle dataset via a set of read-only REST endpoints under `/api/v1/football`.

### Development Plan

| Phase | Status | Description |
|-------|--------|-------------|
| **Phase 1 — MVP** | ✅ Implemented | Schema, import script, Teams & Matches endpoints, Goals, Shootouts, Head-to-Head, Player goals |
| **Phase 2 — Advanced** | 🔲 Planned | Full-text search, team statistics (win/draw/loss), tournament bracket views, pagination cursors |

### Database Schema

The football schema lives in `migrations/002_football_schema.sql`.  Apply it
after the initial migration:

```bash
psql "$DATABASE_URL" -f migrations/002_football_schema.sql
```

Tables created:

| Table | Description |
|-------|-------------|
| `football_teams` | Unique national team names |
| `football_tournaments` | Unique competition names |
| `football_matches` | Match results from `results.csv` |
| `football_goalscorers` | Individual goal events from `goalscorers.csv` |
| `football_shootouts` | Penalty-shootout winners from `shootouts.csv` |
| `football_former_names` | Historical team names from `former_names.csv` |

### Importing the Dataset

The import script downloads the Kaggle ZIP, extracts the four CSV files, and
loads them into the database inside a single transaction (idempotent — safe to
re-run).

**Prerequisites:**

* `DATABASE_URL` pointing to a PostgreSQL instance with the football schema applied.
* Either:
  * Kaggle API credentials (`KAGGLE_USERNAME` + `KAGGLE_KEY`), **or**
  * The ZIP archive placed at `./football_data.zip` to skip the download.

```bash
# With Kaggle credentials
DATABASE_URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable" \
KAGGLE_USERNAME=your_username \
KAGGLE_KEY=your_api_key \
go run scripts/import_football_data.go

# With a pre-downloaded ZIP
DATABASE_URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable" \
cp /path/to/archive.zip ./football_data.zip && \
go run scripts/import_football_data.go
```

The script logs progress for each step and prints a summary on completion.

### Football API Endpoints

`GET` endpoints are public and require no authentication. `POST`, `PUT`, and `DELETE` endpoints are protected and require a valid JWT (see [Authentication](#authentication)).

Base URL: `http://localhost:8080/api/v1/football`

#### Teams

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/teams` | — | List all national teams (alphabetical order) |
| `GET` | `/teams/:id` | — | Get a single team by ID |
| `GET` | `/teams/:id/history` | — | Get the historical names for a team |
| `POST` | `/teams` | JWT | Create a new team |
| `PUT` | `/teams/:id` | JWT | Update an existing team |
| `DELETE` | `/teams/:id` | JWT | Delete a team |

#### Matches

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/matches` | — | List matches (paginated; `?limit=50&offset=0`) |
| `GET` | `/matches/:id` | — | Get a single match by ID |
| `GET` | `/matches/:id/goals` | — | Get all goals scored in a match |
| `GET` | `/matches/:id/shootout` | — | Get the penalty-shootout result for a match (404 if none) |
| `GET` | `/head-to-head?teamA=:id&teamB=:id` | — | Get all matches between two teams |
| `POST` | `/matches` | JWT | Create a new match |
| `PUT` | `/matches/:id` | JWT | Update an existing match |
| `DELETE` | `/matches/:id` | JWT | Delete a match |
| `POST` | `/matches/:id/goals` | JWT | Add a goal to a match |
| `DELETE` | `/matches/:id/goals/:goalId` | JWT | Remove a goal from a match |
| `POST` | `/matches/:id/shootout` | JWT | Record the penalty-shootout result for a match |
| `DELETE` | `/matches/:id/shootout` | JWT | Remove the penalty-shootout result for a match |

#### Players

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/players/:name/goals` | — | Get all goals scored by a player (exact name match) |

### Example Requests

```bash
# List all teams
curl http://localhost:8080/api/v1/football/teams

# Get team with ID 1
curl http://localhost:8080/api/v1/football/teams/1

# Get historical names for team 1
curl http://localhost:8080/api/v1/football/teams/1/history

# List first 20 matches starting from offset 100
curl "http://localhost:8080/api/v1/football/matches?limit=20&offset=100"

# Get goals for match 42
curl http://localhost:8080/api/v1/football/matches/42/goals

# Get penalty-shootout result for match 42
curl http://localhost:8080/api/v1/football/matches/42/shootout

# Head-to-head between teams 1 and 2
curl "http://localhost:8080/api/v1/football/head-to-head?teamA=1&teamB=2"

# All goals scored by a player
curl http://localhost:8080/api/v1/football/players/Ronaldo/goals
```

### Example Response

**GET /api/v1/football/matches/1**

```json
{
  "id": 1,
  "date": "1872-11-30T00:00:00Z",
  "homeTeam": "Scotland",
  "awayTeam": "England",
  "homeTeamId": 45,
  "awayTeamId": 12,
  "homeScore": 0,
  "awayScore": 0,
  "tournament": "Friendly",
  "tournamentId": 3,
  "city": "Glasgow",
  "country": "Scotland",
  "neutral": false,
  "links": [
    {"rel": "self",     "href": "/api/v1/football/matches/1",          "method": "GET"},
    {"rel": "goals",    "href": "/api/v1/football/matches/1/goals",    "method": "GET"},
    {"rel": "shootout", "href": "/api/v1/football/matches/1/shootout", "method": "GET"}
  ]
}
```
