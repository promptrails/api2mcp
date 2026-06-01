# api2mcp

> Turn an existing HTTP API into an MCP server. OpenAPI-first, framework-agnostic, **safe by default**.

## What is api2mcp?

`api2mcp` exposes an existing Go HTTP API to LLMs as a set of [Model Context
Protocol](https://modelcontextprotocol.io) tools — without hand-writing a tool
per endpoint. Point it at an OpenAPI spec (or a live Gin/Echo/Fiber/chi
router), curate which operations are safe to expose, and serve over stdio or
streamable-HTTP.

```go
src, _ := openapi.FromFile("openapi.yaml")
srv := api2mcp.New(src,
    api2mcp.WithBaseURL("https://api.internal"),
    api2mcp.ReadOnly(),            // only GET/HEAD become tools
    api2mcp.IncludeTags("public"), // whitelist
)
srv.ServeStdio(context.Background())
```

## Why another one?

The existing Go options split into two camps: framework-specific magic
(`gin-mcp`, `echo-mcp` — one framework each) and bare OpenAPI→MCP proxies.
`api2mcp` is neither. Its bet is that the hard, valuable part isn't *converting*
an API — it's **curating** it: exposing only what's safe to hand an LLM,
read-only by default, with response shaping so a huge JSON payload doesn't blow
up the model's context.

## Features

| Feature | Description |
|---------|-------------|
| **OpenAPI-first** | Any OpenAPI 3 spec (file or URL) becomes an MCP server |
| **swaggo** | swaggo-annotated apps (OpenAPI 2.0) supported via auto-conversion |
| **Framework adapters** | Embed MCP inside a live Gin, Echo, Fiber or chi app |
| **Curation** | Read-only mode, tag/path/operation include & exclude filters |
| **Safety hints** | Auto MCP tool annotations (read-only / destructive / idempotent) |
| **Auth forwarding** | Pass the caller's Bearer/JWT through to the upstream API |
| **Response shaping** | Cap payload size so large responses don't flood context |
| **Transports** | stdio (desktop clients) and streamable-HTTP (hosted) |
| **CLI** | A standalone binary driven by a YAML config — no Go code required |

## Install

```bash
go get github.com/promptrails/api2mcp
# standalone binary:
go install github.com/promptrails/api2mcp/cmd/api2mcp@latest
```

Requires Go 1.22+.

## Quick Links

- [Getting Started](getting-started.md)
- [Architecture](architecture.md)
- [Curation & Safety](curation.md)
- [GitHub Repository](https://github.com/promptrails/api2mcp)
