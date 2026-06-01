# Framework Adapters

Adapters power **embedded mode**: expose a live router as a Source so you can
mount an MCP endpoint inside an existing app — no OpenAPI spec, no separate
process. api2mcp ships adapters for the four most common Go routers.

| Router | Package |
|--------|---------|
| Gin | `adapter/ginadapter` |
| Echo | `adapter/echoadapter` |
| Fiber | `adapter/fiberadapter` |
| chi | `adapter/chiadapter` |

## Usage

```go
import (
    "github.com/promptrails/api2mcp"
    "github.com/promptrails/api2mcp/adapter/ginadapter"
)

src := ginadapter.New(router)
srv := api2mcp.New(src,
    api2mcp.WithBaseURL("http://localhost:8080"),
    api2mcp.ReadOnly(),
    api2mcp.WithEndpointPath("/mcp"),
)
h, _ := srv.HTTPHandler(context.Background())
router.Any("/mcp", gin.WrapH(h))   // mount MCP inside the same app
```

Echo, Fiber and chi are identical apart from the constructor:

```go
echoadapter.New(e)        // *echo.Echo
fiberadapter.New(app)     // *fiber.App
chiadapter.New(r)         // chi.Routes (e.g. *chi.Mux)
```

## What adapters can and can't see

A router knows each route's **method** and **path** — and that's it. So adapters
emit operations with:

- the HTTP method,
- the path template (router syntax like `:id` / `*x` is normalized to `{id}`),
- required path parameters extracted from the template.

A router does **not** carry request/response body schemas. So a `POST` tool
produced by an adapter alone has no body schema. Two ways to fill that gap:

- **[swaggo](swaggo.md)** — if your handlers are annotated, use the swaggo
  source instead and skip adapters entirely (richest schemas).
- **[Schema providers](schema-providers.md)** — register a Go request struct
  per operation and api2mcp reflects it into a body schema.

> **Recommendation:** in embedded mode, pair adapters with `ReadOnly()`. GET
> endpoints rarely need a body schema, so a read-only embedded server is useful
> immediately; add reflection only when you want to expose writes.

## Tool naming

Tool names are synthesized from method + path, e.g. `GET /users/{id}` becomes
`get_users_id`. Names are sanitized to `[A-Za-z0-9_]` so every MCP client
handles them reliably.
