# Curation & Safety

This is the point of api2mcp. Converting an API to tools is commodity; deciding
**which** operations an LLM may call, and how, is where the risk lives. Curation
is safe by default and composable.

## Read-only by default

The single most important switch. `ReadOnly()` drops every mutating operation
(POST/PUT/PATCH/DELETE), so an LLM physically cannot write through your API:

```go
srv := api2mcp.New(src,
    api2mcp.WithBaseURL("https://api.internal"),
    api2mcp.ReadOnly(),
)
```

When pointing an LLM at a real API for the first time, start here.

## Include / exclude filters

Filters compose. Includes are a **whitelist** (an operation must match at least
one include to survive); excludes are a **blacklist** (matching any exclude
drops it). Excludes win over includes.

```go
srv := api2mcp.New(src,
    api2mcp.WithBaseURL("https://api.internal"),

    // whitelist
    api2mcp.IncludeTags("public", "read"),
    api2mcp.IncludePaths("/v1/*"),

    // blacklist (belt and suspenders)
    api2mcp.ExcludePaths("/admin/*", "/internal/*"),
    api2mcp.ExcludeOperations("deleteEverything"),
)
```

| Option | Keeps / drops by |
|--------|------------------|
| `ReadOnly()` | drop non-GET/HEAD |
| `IncludeTags(...)` / `ExcludeTags(...)` | OpenAPI tag |
| `IncludePaths(...)` / `ExcludePaths(...)` | path glob (`/admin/*` matches recursively) |
| `IncludeOperations(...)` / `ExcludeOperations(...)` | exact operation id |
| `WithFilter(fn)` | arbitrary `func(ir.Operation) bool` |

### Custom predicates

For anything the built-ins don't cover:

```go
// only expose operations that require no auth
api2mcp.WithFilter(func(op ir.Operation) bool {
    return len(op.Security) == 0
})
```

## Filtering order

For each operation, the policy applies in this order:

1. if `ReadOnly` is set and the op is not read-only → **drop**;
2. if any includes exist and none match → **drop**;
3. if any exclude matches → **drop**;
4. otherwise → **keep**.

## Safety hints (automatic)

Every generated tool carries [MCP tool annotations](https://modelcontextprotocol.io/docs/concepts/tools)
derived from its HTTP method. MCP clients surface these to the user — e.g.
warning or asking for confirmation before invoking a destructive tool — so they
are a real safety signal, set for you automatically:

| Method | `readOnlyHint` | `destructiveHint` | `idempotentHint` |
|--------|:---:|:---:|:---:|
| GET / HEAD | ✓ | ✗ | ✓ |
| POST | ✗ | ✗ | ✗ |
| PUT | ✗ | ✓ | ✓ |
| PATCH | ✗ | ✓ | ✗ |
| DELETE | ✗ | ✓ | ✓ |

`openWorldHint` is always true (a tool call reaches an external API). No
configuration is required — this is on by default.

## Audit logging

Make LLM-driven traffic observable: register a callback invoked after every tool
call with the operation, upstream status, duration and any error.

```go
srv := api2mcp.New(src,
    api2mcp.WithBaseURL("https://api.internal"),
    api2mcp.WithAuditLogger(api2mcp.StdAuditLogger), // one log line per call
)
```

`StdAuditLogger` emits:

```
api2mcp: getUser GET /users/{id} -> 200 (12ms)
```

For custom sinks (structured logs, metrics, tracing) pass your own
`func(engine.AuditEvent)`:

```go
api2mcp.WithAuditLogger(func(e engine.AuditEvent) {
    metrics.Observe(e.OperationID, e.Status, e.Duration)
})
```

In the CLI, set `audit: true` in the config.

## Recommended starting posture

```go
api2mcp.New(src,
    api2mcp.WithBaseURL("https://api.internal"),
    api2mcp.ReadOnly(),               // no writes
    api2mcp.IncludeTags("ai-safe"),   // explicit opt-in per endpoint
    api2mcp.ExcludePaths("/internal/*"),
    api2mcp.WithMaxResponseBytes(16 << 10), // don't flood context
    api2mcp.WithAuditLogger(api2mcp.StdAuditLogger), // observe every call
)
```

Then loosen deliberately, one tag or operation at a time.

## Next

- [Transports & Auth](transports.md) — response shaping and auth forwarding
