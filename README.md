# api2mcp

Turn an existing HTTP API into an [MCP](https://modelcontextprotocol.io) server — **OpenAPI-first**, framework-agnostic, and **curation-first** (safe by default).

## Why another one?

The existing Go options split into two camps: framework-specific magic
(`gin-mcp`, `echo-mcp` — low traction, one framework each) and bare
OpenAPI→MCP proxies. `api2mcp` is neither a single-framework toy nor a blind
proxy. Its bet is that the hard, valuable part isn't *converting* an API — it's
**curating** it: exposing only what's safe to hand an LLM, read-only by
default, with response shaping so a 200-field JSON payload doesn't blow up the
model's context.

It is **layered** so each concern is replaceable:

```
L5  Transport      stdio · streamable-HTTP                (mark3labs/mcp-go)
L4  Curation       include/exclude · read-only · shape
L3  Engine         IR → MCP tool + HTTP executor          (framework-agnostic core)
L2  Sources        OpenAPI · swaggo · framework adapters
L1  Adapters       gin · echo · fiber · chi   (+ reflect schema provider)
```

Everything collapses into one intermediate representation (`ir.Operation`), so
the core has exactly one code path whether the input is a swagger.json or a
live Gin router.

## Install

```bash
go get github.com/promptrails/api2mcp
# or the standalone binary:
go install github.com/promptrails/api2mcp/cmd/api2mcp@latest
```

## Quick start

**Library — OpenAPI spec over stdio (Claude Desktop, Cursor, Zed):**

```go
src, _ := openapi.FromFile("openapi.yaml")
srv := api2mcp.New(src,
    api2mcp.WithBaseURL("https://api.internal"),
    api2mcp.ReadOnly(),              // only GET/HEAD become tools
    api2mcp.IncludeTags("public"),   // whitelist
)
srv.ServeStdio(context.Background())
```

**CLI — no Go code, just a config file:**

```bash
api2mcp serve -config examples/api2mcp.yaml
# or the quick path:
api2mcp serve -openapi openapi.yaml -base https://api.internal -read-only -http :8080
```

**Embedded — mount /mcp inside an existing Gin/Echo/Fiber/chi app:**

```go
src := ginadapter.New(router)        // routes are the source of tools
srv := api2mcp.New(src, api2mcp.WithBaseURL("http://localhost:8080"), api2mcp.ReadOnly())
h, _ := srv.HTTPHandler(context.Background())
router.Any("/mcp", gin.WrapH(h))
```

## Sources

| Source | Use it when |
|---|---|
| `source/openapi` | You have an OpenAPI 3 spec (file or URL). **Primary path.** |
| `source/swaggo` | Your app is annotated with [swaggo](https://github.com/swaggo/swag) (OpenAPI 2.0). |
| `adapter/{gin,echo,fiber,chi}` | Embedded mode: the live router is the source. |
| `schema/reflectschema` | Embedded mode + you want request-body schemas from Go structs. |

## Curation (the point)

Safe by default. Compose any of:

- `ReadOnly()` — drop every mutating operation (POST/PUT/PATCH/DELETE).
- `IncludeTags(...)` / `ExcludeTags(...)`
- `IncludePaths("/public/*")` / `ExcludePaths("/admin/*")`
- `IncludeOperations(id...)` / `ExcludeOperations(id...)`
- `WithFilter(func(ir.Operation) bool)` — arbitrary predicate.

Plus:

- Every tool gets **MCP safety annotations**
  (`readOnlyHint`/`destructiveHint`/`idempotentHint`) derived from its HTTP
  method, so clients can warn before destructive calls.
- `WithMarkDestructive("charge")` — force the destructive hint on risky endpoints.
- `WithRateLimit(5, 10)` — per-tool token bucket, protects the upstream from a runaway LLM.
- `WithAuditLogger(api2mcp.StdAuditLogger)` — log every tool call.

## Transport & auth

- `ServeStdio(ctx)` for desktop clients; `ServeHTTP(ctx, ":8080")` for hosted/streamable-HTTP.
- `WithForwardHeaders("Authorization")` passes the caller's Bearer/JWT through to the upstream API.
- `WithStaticHeader(k, v)` injects a fixed header (e.g. a server-side API key).
- `WithMaxResponseBytes(n)` caps each response so large payloads don't flood the model's context.

## Examples

- `examples/openapi-stdio` — OpenAPI → stdio
- `examples/openapi-http` — OpenAPI → streamable-HTTP with auth forwarding
- `examples/embedded-gin` — MCP mounted inside a Gin app
- `examples/embedded-echo` — MCP mounted inside an Echo app
- `examples/embedded-fiber` — MCP mounted inside a Fiber app
- `examples/embedded-chi` — MCP mounted inside a chi app
- `examples/api2mcp.yaml` — full CLI config

All four embedded examples mount `/mcp` in the same process as the API and are
verified end-to-end (an MCP `tools/call` reaches the app's own handler).

## License

MIT
