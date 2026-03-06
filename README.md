# COMP3011 Coursework 1 — International Football Results API (Gin / Go + PostgreSQL)

A production-ready RESTful API in Go using the
[Gin](https://github.com/gin-gonic/gin) framework backed by **PostgreSQL**
(via [`lib/pq`](https://github.com/lib/pq)).  The API exposes data from the
[International football results from 1872 to 2025](https://www.kaggle.com/datasets/martj42/international-football-results-from-1872-to-2017)
Kaggle dataset.  Every architectural decision is explicitly mapped to one or
more of the **Six Guiding Principles of REST** (Fielding, 2000).

A PostgreSQL database is **required** to run the server.

---

## Six Guiding Principles of REST

| # | Principle | How this project addresses it |
|---|-----------|-------------------------------|
| 1 | **Client–Server** | HTTP handlers (`internal/handlers`) are completely decoupled from any client implementation. The server exposes a uniform HTTP interface; clients are free to be web browsers, mobile apps, or CLI tools. |
| 2 | **Stateless** | No server-side session state exists. Every request must carry all information needed to be processed (validated body / path parameters). Authentication is token-based using JWT — all user identity is carried in the self-describing token, not in server-side sessions. |
| 3 | **Cacheable** | The `CacheControl` middleware sets `Cache-Control: public, max-age=60` on `GET`/`HEAD` responses, enabling clients and intermediary caches to store them. Mutating methods (`POST`, `PUT`, `DELETE`) are marked `no-store`. |
| 4 | **Uniform Interface** | Resources are identified by versioned URIs (e.g. `/api/v1/football/teams/{id}`). Standard HTTP verbs (`GET`, `POST`, `PUT`, `DELETE`) map to CRUD operations. Response bodies include HATEOAS hypermedia links so clients can discover related actions without out-of-band knowledge. |
| 5 | **Layered System** | Middleware (`RequestID`, `Logger`, `JWTAuth`, `CacheControl`, `Recovery`) forms transparent processing layers between the network and the handlers. The same binary runs correctly behind a load-balancer or reverse proxy. |
| 6 | **Code on Demand** *(optional)* | Not implemented by default. The architecture supports it — a handler could return executable JavaScript or WebAssembly to extend client functionality when required. |

---

## Project Layout

```
.
├── cmd/
│   └── server/
│       └── main.go                  # Entry point — reads PORT, JWT_SECRET, DATABASE_URL env vars
├── internal/
│   ├── auth/
│   │   └── jwt.go                   # JWT token generation and validation
│   ├── db/
│   │   ├── repository.go            # Repository interfaces (FootballRepository, UserRepository)
│   │   └── postgres/
│   │       ├── db.go                # PostgreSQL connection helper (Connect / ConnectFromEnv)
│   │       ├── football_repo.go     # PostgreSQL FootballRepo — implements FootballRepository
│   │       └── user_repo.go         # PostgreSQL UserRepo — implements UserRepository
│   ├── handlers/
│   │   ├── auth.go                  # Authentication endpoints (register, login)
│   │   ├── football_handler.go      # FootballHandler + shared helpers (HATEOAS links)
│   │   ├── football_teams.go        # Teams CRUD handlers
│   │   ├── football_matches.go      # Matches CRUD handlers
│   │   ├── football_goals.go        # Goals & Shootouts handlers
│   │   ├── football_teams_test.go   # Teams handler tests
│   │   ├── football_matches_test.go # Matches handler tests
│   │   └── football_goals_test.go   # Goals & Shootouts handler tests
│   ├── middleware/
│   │   ├── auth.go                  # JWT authentication middleware
│   │   └── middleware.go            # RequestID, Logger, CacheControl, NoSessionState
│   ├── models/
│   │   ├── common.go                # Shared types: Link, ErrorResponse
│   │   ├── errors.go                # Shared sentinel errors (ErrNotFound, ErrConflict)
│   │   ├── match.go                 # Match, Goal, Shootout domain models
│   │   ├── team.go                  # Team, FormerName domain models
│   │   ├── tournament.go            # Tournament domain model
│   │   └── user.go                  # User domain model + auth request/response types
│   └── router/
│       └── router.go                # Wires middleware, repositories, and routes together
├── migrations/
│   ├── 001_initial_schema.sql       # Idempotent DDL — users table
│   ├── 002_football_schema.sql      # Idempotent DDL — football tables + indexes
│   └── 003_drop_items_table.sql     # Drops the obsolete items table (existing databases)
├── scripts/
│   └── import_football_data.go      # Kaggle dataset importer
├── docker-compose.yml               # PostgreSQL 16 for local development
├── go.mod
├── go.sum
└── README.md
```

---

## Getting Started

### Prerequisites

* Go 1.24 or later
* **PostgreSQL 14+** — required to run the server
* **Docker & Docker Compose** *(optional — for the quickest local database setup)*

### Option A — run with Docker Compose (recommended)

Create a `.env` file in the project root:

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

# Apply schemas
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql
psql "$DATABASE_URL" -f migrations/002_football_schema.sql

# If upgrading an existing database that has the items table, drop it
psql "$DATABASE_URL" -f migrations/003_drop_items_table.sql

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

### Option B — run with an existing PostgreSQL instance

```bash
# 1. Apply the schemas (safe to run multiple times — all statements are IF NOT EXISTS)
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql
psql "$DATABASE_URL" -f migrations/002_football_schema.sql

# 1a. If upgrading an existing database that has the items table, drop it
psql "$DATABASE_URL" -f migrations/003_drop_items_table.sql

# 2. (Optional) Import the Kaggle dataset
DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" \
  go run scripts/import_football_data.go

# 3. Start the server
DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" \
  JWT_SECRET=your-secret-key \
  go run ./cmd/server
```

### Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | **Yes** (production) | — | Secret key used to sign JWT tokens. Set `DEV_MODE=true` to auto-generate a random one for development (not safe for production). |
| `DATABASE_URL` | **Yes** | — | libpq connection string for PostgreSQL (e.g. `postgres://user:pass@host:5432/dbname?sslmode=disable`) |
| `PORT` | No | `8080` | TCP port the server listens on |
| `DEV_MODE` | No | — | Set to `true` to auto-generate `JWT_SECRET` in development |

### Run the tests

```bash
go test ./...
```

The handler tests use in-process mock repositories, so no database connection is required to run the test suite.

### Build a binary

```bash
go build -o api-server ./cmd/server
DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" \
  JWT_SECRET=your-secret-key \
  ./api-server
```

---

## Database

### Schemas

Three migration files live under `migrations/`.  Every statement uses
`IF NOT EXISTS` / `IF EXISTS`, so the files are safe to re-apply.

#### `migrations/001_initial_schema.sql` — users

```sql
-- users: bcrypt-hashed passwords only — plain text never stored
CREATE TABLE IF NOT EXISTS users (
    username      VARCHAR(50)  PRIMARY KEY,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

#### `migrations/002_football_schema.sql` — football data

```sql
CREATE TABLE IF NOT EXISTS football_teams (
    id         SERIAL       PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS football_tournaments (
    id         SERIAL       PRIMARY KEY,
    name       VARCHAR(200) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS football_matches (
    id            SERIAL       PRIMARY KEY,
    match_date    DATE         NOT NULL,
    home_team_id  INTEGER      NOT NULL REFERENCES football_teams(id),
    away_team_id  INTEGER      NOT NULL REFERENCES football_teams(id),
    home_score    INTEGER      NOT NULL,
    away_score    INTEGER      NOT NULL,
    tournament_id INTEGER      NOT NULL REFERENCES football_tournaments(id),
    city          VARCHAR(100) NOT NULL DEFAULT '',
    country       VARCHAR(100) NOT NULL DEFAULT '',
    neutral       BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (match_date, home_team_id, away_team_id)
);

CREATE TABLE IF NOT EXISTS football_goalscorers (
    id         SERIAL       PRIMARY KEY,
    match_id   INTEGER      NOT NULL REFERENCES football_matches(id),
    team_id    INTEGER      NOT NULL REFERENCES football_teams(id),
    scorer     VARCHAR(100) NOT NULL,
    own_goal   BOOLEAN      NOT NULL DEFAULT FALSE,
    penalty    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS football_shootouts (
    id         SERIAL      PRIMARY KEY,
    match_id   INTEGER     NOT NULL REFERENCES football_matches(id) UNIQUE,
    winner_id  INTEGER     NOT NULL REFERENCES football_teams(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS football_former_names (
    id          SERIAL       PRIMARY KEY,
    team_id     INTEGER      NOT NULL REFERENCES football_teams(id),
    former_name VARCHAR(100) NOT NULL,
    start_date  DATE,
    end_date    DATE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

#### `migrations/003_drop_items_table.sql` — remove obsolete items table

Apply this on any existing database that was provisioned before the Item
resource was removed:

```sql
DROP INDEX IF EXISTS items_name_idx;
DROP INDEX IF EXISTS items_updated_at_idx;
DROP TABLE IF EXISTS items;
```

#### `migrations/005_elo_ratings.sql` — Elo rating cache and config

Adds the `football_elo_cache` (pre-computed Elo snapshots) and
`football_elo_config` (runtime-tunable parameters) tables:

```sql
CREATE TABLE IF NOT EXISTS football_elo_cache (
    id              SERIAL PRIMARY KEY,
    team_id         INTEGER NOT NULL REFERENCES football_teams(id) ON DELETE CASCADE,
    as_of_date      DATE NOT NULL,
    elo_rating      NUMERIC(8,2) NOT NULL,
    global_rank     INTEGER,
    matches_played  INTEGER NOT NULL DEFAULT 0,
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (team_id, as_of_date)
);

CREATE TABLE IF NOT EXISTS football_elo_config (
    key         VARCHAR(50) PRIMARY KEY,
    value       JSONB NOT NULL,
    description TEXT,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Connection pooling

`internal/db/postgres/db.go` configures the `*sql.DB` pool:

| Setting | Value | Rationale |
|---------|-------|-----------|
| `MaxOpenConns` | 25 | Hard cap on concurrent DB connections |
| `MaxIdleConns` | 5 | Keep a small warm pool to reduce connection-setup latency |
| `ConnMaxLifetime` | 5 min | Recycle connections before load-balancer / firewall idle limits are hit |

### Repository pattern

Repository interfaces are declared in `internal/db/repository.go`:

```go
type FootballRepository interface { ... }
type UserRepository     interface { ... }
```

`router.New` receives a `*sql.DB` and wires in the PostgreSQL implementations
(`postgres.NewFootballRepo`, `postgres.NewUserRepo`).  Handlers depend only on
these interfaces, making them easy to test with mock implementations.

---

## Importing the Dataset

The import script loads the Kaggle ZIP, extracts the four CSV files, and
loads them into the database inside a single transaction (idempotent — safe to
re-run).

**Prerequisites:**

* `DATABASE_URL` pointing to a PostgreSQL instance with the football schema applied.
* The ZIP archive placed at `./football_data.zip`.

```bash
cp /path/to/archive.zip ./football_data.zip
DATABASE_URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable" \
go run scripts/import_football_data.go
```

The script logs progress for each step and prints a summary on completion.

---

## API Reference

Base URL: `http://localhost:8080/api/v1`

### Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/auth/register` | — | Register a new user account |
| `POST` | `/auth/login` | — | Login and receive a JWT token |

### Football — Teams

`GET` endpoints are public. `POST`, `PUT`, and `DELETE` endpoints require a valid JWT.

Base path: `/api/v1/football`

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/teams` | — | List all national teams (alphabetical order) |
| `GET` | `/teams/:id` | — | Get a single team by ID |
| `GET` | `/teams/:id/history` | — | Get the historical names for a team |
| `POST` | `/teams` | JWT | Create a new team |
| `PUT` | `/teams/:id` | JWT | Update an existing team |
| `DELETE` | `/teams/:id` | JWT | Delete a team |

### Football — Matches

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

### Football — Players

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/players/:name/goals` | — | Get all goals scored by a player (exact name match) |

### Football — Elo Ratings

Dynamic World Football Elo ratings computed from historical match data (1872–present).
See [docs/elo-methodology.md](docs/elo-methodology.md) for the full formula and parameter details.

| Method | Path | Auth | Description | Query Params |
|--------|------|------|-------------|--------------|
| `GET` | `/teams/:id/elo` | — | Get current or historical Elo rating for a team | `?date=YYYY-MM-DD`, `?include_history=true` |
| `GET` | `/teams/:id/elo/timeline` | — | Time-series of Elo changes for a team | `?start_date=`, `?end_date=`, `?resolution=match\|month\|year` |
| `GET` | `/rankings/elo` | — | Global Elo rankings snapshot (returns empty + `X-Cache-Status: miss` if cache not pre-warmed) | `?date=YYYY-MM-DD`, `?region=europe\|asia\|…`, `?limit=50&offset=0` |
| `POST` | `/rankings/elo/recalculate` | JWT | Trigger background Elo recalculation (admin). Rate-limited: once per 5 min; use `?force=true` to bypass. Returns 429 if already running. | `?team_id=optional`, `?force=true` |

**Elo response example** (`GET /teams/45/elo?date=2014-07-13`):

```json
{
  "teamId": 45,
  "teamName": "Germany",
  "date": "2014-07-13T00:00:00Z",
  "elo": 2145.30,
  "changeFromPrevious": 12.40,
  "matchesConsidered": 847,
  "methodology": {
    "kFactor": 5,
    "homeAdvantage": 100,
    "weightMultiplier": 1.0,
    "formulaReference": "https://www.eloratings.net/method.html"
  },
  "links": [
    {"rel": "self",     "href": "/api/v1/football/teams/45/elo?date=2014-07-13", "method": "GET"},
    {"rel": "timeline", "href": "/api/v1/football/teams/45/elo/timeline",        "method": "GET"},
    {"rel": "team",     "href": "/api/v1/football/teams/45",                     "method": "GET"}
  ]
}
```

**Elo environment variables:**

| Variable | Default | Description |
|----------|---------|-------------|
| `ELO_DEFAULT_RATING` | `1500` | Starting Elo for new teams |
| `ELO_HOME_ADVANTAGE` | `100` | Points added to home-team expected result |
| `ELO_GOAL_MARGIN_FACTOR` | `0.1` | Coefficient for `ln(\|goal_diff\|+1)` adjustment |

### Response Headers

| Header | Description |
|--------|-------------|
| `X-Request-ID` | Unique ID for each request (traceability) |
| `Cache-Control` | `public, max-age=60` on GET; `no-store` on mutations and the recalculate endpoint |
| `Location` | Set to the new resource URI on `201 Created` |
| `X-Elo-Computed-At` | Timestamp of when the Elo rating was computed (Elo endpoints only) |
| `X-Cache-Status` | `hit` or `miss` on `GET /rankings/elo`; `miss` means no snapshot exists for the date — pre-warm with `/recalculate` |

---

## Request / Response Examples

### Authentication

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
    {"rel":"football","href":"/api/v1/football/teams","method":"GET"}
  ]
}
```

### Football

**List all teams**

```bash
curl http://localhost:8080/api/v1/football/teams
```

**Get a single team**

```bash
curl http://localhost:8080/api/v1/football/teams/1
```

**Get historical names for a team**

```bash
curl http://localhost:8080/api/v1/football/teams/1/history
```

**List matches (paginated)**

```bash
curl "http://localhost:8080/api/v1/football/matches?limit=20&offset=100"
```

**Get goals for a match**

```bash
curl http://localhost:8080/api/v1/football/matches/42/goals
```

**Get penalty-shootout result for a match**

```bash
curl http://localhost:8080/api/v1/football/matches/42/shootout
```

**Head-to-head between two teams**

```bash
curl "http://localhost:8080/api/v1/football/head-to-head?teamA=1&teamB=2"
```

**All goals scored by a player**

```bash
curl http://localhost:8080/api/v1/football/players/Ronaldo/goals
```

**Create a team** (requires JWT)

```bash
curl -X POST http://localhost:8080/api/v1/football/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"name":"New Team"}'
```

**Example match response**

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
    {"rel": "update",   "href": "/api/v1/football/matches/1",          "method": "PUT"},
    {"rel": "delete",   "href": "/api/v1/football/matches/1",          "method": "DELETE"},
    {"rel": "goals",    "href": "/api/v1/football/matches/1/goals",    "method": "GET"},
    {"rel": "shootout", "href": "/api/v1/football/matches/1/shootout", "method": "GET"}
  ]
}
```

---

## Extending the Project

1. **Add a new resource** — create a handler file in `internal/handlers/`,
   define a new repository interface in `internal/db/repository.go`, implement
   it in `internal/db/postgres/`, and add routes in `internal/router/router.go`.
2. **Add a database migration** — create the next numbered SQL file in
   `migrations/` (e.g. `003_add_statistics.sql`) and apply it with `psql`.
3. **Add role-based access control** — extend the JWT claims in
   `internal/auth/jwt.go` to include roles, then add middleware to check
   permissions before reaching handlers.
4. **Add pagination** — indexes on `football_matches` already support efficient
   `LIMIT`/`OFFSET` or cursor-based queries; extend `ListMatches` in
   `internal/db/postgres/football_repo.go` and the route handler.
5. **Configuration** — read additional settings from environment variables or
   a config file in `cmd/server/main.go`.
