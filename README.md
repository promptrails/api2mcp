# api2mcp

Turn an existing HTTP API into an [MCP](https://modelcontextprotocol.io) server — **OpenAPI-first**, framework-agnostic, and **curation-first** (safe by default).

> Status: building in milestones. M0 (OpenAPI → MCP over stdio) works today.

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
L5  Transport      stdio · streamable-HTTP · SSE        (mark3labs/mcp-go)
L4  Curation       include/exclude · read-only · shape
L3  Engine         IR → MCP tool + HTTP executor        (framework-agnostic core)
L2  Sources        OpenAPI spec  ·  framework adapters
L1  Adapters       gin · echo · fiber · chi   (+ swaggo/reflect schema providers)
```

Everything collapses into one intermediate representation (`ir.Operation`), so
the core has exactly one code path whether the input is a swagger.json or a
live Gin router.

## Quick start (M0)

```go
src, _ := openapi.FromFile("openapi.yaml")
srv := api2mcp.New(src, api2mcp.WithBaseURL("https://api.internal"))
srv.ServeStdio(context.Background())
```

Run the bundled example against a public API:

```bash
go run ./examples/openapi-stdio -spec examples/openapi-stdio/openapi.yaml
```

## Milestones

- [x] **M0** — OpenAPI → MCP tools over stdio (working spike)
- [x] **M1** — Curation: include/exclude filters, read-only mode
- [x] **M2** — streamable-HTTP transport, auth forwarding, response shaping
- [ ] **M3** — Framework adapters (gin, echo, fiber, chi) for embedded mode
- [ ] **M4** — Schema providers (swaggo, struct reflection)
- [ ] **M5** — CLI binary, config file, docs

## License

MIT
