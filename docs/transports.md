# Transports & Auth

## Choosing a transport

| Transport | Use when | Method |
|-----------|----------|--------|
| **stdio** | Desktop clients (Claude Desktop, Cursor, Zed) launch your server as a subprocess | `ServeStdio(ctx)` |
| **streamable-HTTP** | Hosted / remote, behind a load balancer | `ServeHTTP(ctx, ":8080")` |

```go
// stdio
srv.ServeStdio(context.Background())

// streamable-HTTP
srv.ServeHTTP(context.Background(), ":8080")
```

The HTTP endpoint path defaults to `/mcp`; override with
`WithEndpointPath("/custom")`.

### Mounting in your own server

To embed the MCP endpoint inside an existing `http.Server` or router, grab the
handler instead of calling `ServeHTTP`:

```go
h, _ := srv.HTTPHandler(context.Background())
mux.Handle("/mcp", h)            // net/http
router.Any("/mcp", gin.WrapH(h)) // gin
```

## Auth forwarding

When the upstream API is protected, forward the caller's credentials through to
it. `WithForwardHeaders` copies named headers from the incoming MCP HTTP request
onto every upstream call:

```go
srv := api2mcp.New(src,
    api2mcp.WithBaseURL("https://api.internal"),
    api2mcp.WithForwardHeaders("Authorization"), // client's Bearer/JWT → upstream
)
srv.ServeHTTP(ctx, ":8080")
```

This is only meaningful over HTTP (stdio has no per-request headers).

### Static credentials

To inject a fixed, server-side credential (e.g. an API key the client never
sees), use `WithStaticHeader`:

```go
api2mcp.WithStaticHeader("X-Api-Key", os.Getenv("UPSTREAM_API_KEY"))
```

Forwarded headers take precedence over static ones when both set the same key.

## Response shaping

Upstream responses go straight to the model as tool output. A large JSON payload
can blow up the context window, so cap it:

```go
api2mcp.WithMaxResponseBytes(16 << 10) // 16 KiB; truncated with a clear notice
```

When truncated, the output ends with `…[truncated: N of M bytes shown]` so the
model knows data was clipped rather than silently lost.

For full control over rendering, supply your own shaper:

```go
api2mcp.WithResponseShaper(func(r *engine.Response) string {
    // e.g. redact fields, pretty-print, summarize
    return string(r.Body)
})
```

Non-2xx responses are prefixed with `HTTP <status>` by the default shaper so the
LLM can reason about failures.

## Next

- [CLI & Config](cli.md) — run all of this from a YAML file
