# Getting Started

## Installation

```bash
go get github.com/promptrails/api2mcp
```

For the standalone binary:

```bash
go install github.com/promptrails/api2mcp/cmd/api2mcp@latest
```

Requires Go 1.22 or later.

## Three ways to use it

api2mcp meets you wherever your API definition lives.

### 1. OpenAPI spec → stdio

Best for desktop MCP clients (Claude Desktop, Cursor, Zed). Point it at a spec
and serve over stdio:

```go
package main

import (
    "context"
    "log"

    "github.com/promptrails/api2mcp"
    "github.com/promptrails/api2mcp/source/openapi"
)

func main() {
    src, err := openapi.FromFile("openapi.yaml")
    if err != nil {
        log.Fatal(err)
    }
    srv := api2mcp.New(src,
        api2mcp.WithName("my-api"),
        api2mcp.WithBaseURL("https://api.internal"),
        api2mcp.ReadOnly(),              // safe default
        api2mcp.IncludeTags("public"),   // whitelist
    )
    if err := srv.ServeStdio(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### 2. CLI — no Go code

If you just have a spec and a base URL, the binary is enough:

```bash
api2mcp serve -openapi openapi.yaml -base https://api.internal -read-only -http :8080
```

Or drive everything from a YAML file — see [CLI & Config](cli.md).

### 3. Embedded — mount /mcp inside your app

If you run Gin/Echo/Fiber/chi, the live router *is* the source — no spec needed:

```go
src := ginadapter.New(router)
srv := api2mcp.New(src,
    api2mcp.WithBaseURL("http://localhost:8080"),
    api2mcp.ReadOnly(),
    api2mcp.WithEndpointPath("/mcp"),
)
h, _ := srv.HTTPHandler(context.Background())
router.Any("/mcp", gin.WrapH(h))
```

See [Framework Adapters](adapters.md).

## Try the examples

```bash
# OpenAPI over stdio
go run ./examples/openapi-stdio -spec examples/openapi-stdio/openapi.yaml

# OpenAPI over streamable-HTTP
go run ./examples/openapi-http -addr :8080

# MCP embedded inside a Gin app
go run ./examples/embedded-gin
```

## Next steps

- [Architecture](architecture.md) — how the layers fit together
- [Curation & Safety](curation.md) — the point of the library
- [Transports & Auth](transports.md) — stdio vs HTTP, forwarding auth
