# swaggo

[swaggo](https://github.com/swaggo/swag) is the most common way Go services
document their handlers, via `// @Summary`, `// @Param` annotations. It emits a
`docs/swagger.json` in **OpenAPI 2.0**. api2mcp converts that to v3 and reuses
the standard OpenAPI extraction, so an annotated app gets full input schemas
with no adapter and no manual schema work.

## Usage

```go
import "github.com/promptrails/api2mcp/source/swaggo"

src, err := swaggo.FromFile("docs/swagger.json")
// or
src, err := swaggo.FromData(swaggerBytes)

srv := api2mcp.New(src, api2mcp.WithBaseURL("https://api.internal"))
```

The returned value is a regular `*openapi.Source`, so everything in the
[OpenAPI](openapi.md) docs applies — curation, base URL, transports, all work
identically.

## How it works

1. The swaggo `swagger.json` is parsed as an OpenAPI 2.0 document.
2. It is converted to OpenAPI 3 (via `kin-openapi/openapi2conv`).
3. The standard OpenAPI → IR extraction runs on the v3 document.

This means parameters, path templates and request bodies you documented with
swaggo annotations carry straight through into tool schemas.

## When to use swaggo vs adapters

| You have… | Use |
|-----------|-----|
| swaggo annotations + generated `swagger.json` | **swaggo source** (full schemas) |
| A live router, no spec | [Framework adapter](adapters.md) (method+path only) |
| A live router + Go request structs | adapter + [reflection](schema-providers.md) |

If your app is already swaggo-annotated, prefer this source — it gives the
richest tool schemas with the least extra work.
