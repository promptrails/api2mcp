# Architecture

api2mcp is **layered** so each concern is independently replaceable. The MCP
protocol itself is handled by [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go);
api2mcp is the layer above it.

```
L5  Transport      stdio · streamable-HTTP                (mark3labs/mcp-go)
L4  Curation       include/exclude · read-only · shape · safety hints
L3  Engine         IR → MCP tool + HTTP executor          (framework-agnostic core)
L2  Sources        OpenAPI · swaggo · framework adapters
L1  Adapters       gin · echo · fiber · chi   (+ reflect schema provider)
```

## The intermediate representation

Everything collapses into one type, [`ir.Operation`](https://github.com/promptrails/api2mcp/blob/main/ir/ir.go).
Whether the input is a `swagger.json` or a live Gin router, the core sees the
same shape, so there is exactly **one code path** downstream:

```go
type Operation struct {
    ID          string        // becomes the MCP tool name
    Method      string        // GET, POST, ...
    Path        string        // "/users/{id}"
    Summary     string
    Description string
    Tags        []string
    Params      []Param       // path / query / header
    RequestBody *Body         // JSON body schema
    Security    []string
}
```

A `Source` is anything that produces `[]ir.Operation`:

```go
type Source interface {
    Operations(ctx context.Context) ([]ir.Operation, error)
}
```

## Request flow

When an MCP client calls a tool:

1. **Engine** looks up the `Operation` for that tool name.
2. **Executor** builds the upstream HTTP request: substitutes path params,
   appends query params, sets headers, marshals the JSON body.
3. Forwarded headers (e.g. `Authorization`) from the MCP request are attached.
4. The upstream response is **shaped** (status prefix on errors, optional size
   cap) and returned to the LLM as tool output.

## Why this shape

- **Sources are pluggable** — adding a new framework is a ~25-line adapter that
  emits `ir.Operation`; the engine, curation and transport layers are untouched.
- **Curation is centralized** — one policy applies regardless of where
  operations came from, so "read-only" means the same thing for OpenAPI and for
  a live router.
- **The protocol is isolated** — MCP spec churn is absorbed by mcp-go, not
  spread through the codebase.

## Package map

| Package | Layer | Responsibility |
|---------|-------|----------------|
| `ir` | core | The `Operation` intermediate representation |
| `source` | L2 | `Source` interface + helpers |
| `source/openapi` | L2 | OpenAPI 3 → IR |
| `source/swaggo` | L2 | swaggo (OpenAPI 2.0) → IR |
| `source/route` | L1 | Shared route→IR logic for adapters |
| `adapter/{gin,echo,fiber,chi}` | L1 | Live router → Source |
| `schema/reflectschema` | L1 | Struct → request-body schema |
| `engine` | L3 | IR → MCP tool, HTTP executor, shaping |
| `policy` | L4 | Curation matchers + policy |
| `api2mcp` (root) | L4/L5 | Options, wiring, transports |
| `cmd/api2mcp` | — | Standalone CLI |
