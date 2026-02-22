# COMP3011 Coursework 1 — RESTful API Template (Gin / Go)

A production-ready template for building RESTful APIs in Go using the
[Gin](https://github.com/gin-gonic/gin) framework.  Every architectural
decision is explicitly mapped to one or more of the
**Six Guiding Principles of REST** (Fielding, 2000).

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
│       └── main.go          # Entry point — reads PORT and JWT_SECRET env vars
├── internal/
│   ├── auth/
│   │   └── jwt.go           # JWT token generation and validation
│   ├── handlers/
│   │   ├── auth.go          # Authentication endpoints (register, login)
│   │   ├── handlers.go      # Item CRUD handlers + in-memory store
│   │   └── handlers_test.go # Table-driven handler tests
│   ├── middleware/
│   │   ├── auth.go          # JWT authentication middleware
│   │   └── middleware.go    # RequestID, Logger, CacheControl
│   ├── models/
│   │   ├── item.go          # Item domain model + request/response types
│   │   └── user.go          # User domain model + auth request/response types
│   └── router/
│       └── router.go        # Wires middleware and routes together
├── go.mod
├── go.sum
└── README.md
```

---

## Getting Started

### Prerequisites

* Go 1.21 or later

### Run the server

```bash
# Default port 8080, with JWT secret
JWT_SECRET=your-secret-key go run ./cmd/server

# Custom port and JWT secret
JWT_SECRET=your-secret-key PORT=9090 go run ./cmd/server
```

### Run the tests

```bash
go test ./...
```

### Build a binary

```bash
go build -o api-server ./cmd/server
./api-server
```

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

## Extending the Template

1. **Add a new resource** — create a file in `internal/handlers/`, register routes in `internal/router/router.go`.
2. **Replace the in-memory store** — swap `handlers.Store` for a struct that wraps your database connection (e.g., PostgreSQL, MongoDB).
3. **Add role-based access control** — extend the JWT claims in `internal/auth/jwt.go` to include roles, and create middleware to check permissions.
4. **Configuration** — read additional settings from environment variables or a config file in `cmd/server/main.go`.