# CLI & Config

The `api2mcp` binary serves an existing HTTP API as an MCP server with no Go
code — driven by a YAML config or a few flags.

```bash
go install github.com/promptrails/api2mcp/cmd/api2mcp@latest
```

## Quick path (flags)

For an OpenAPI spec, flags alone are enough:

```bash
# stdio (desktop clients)
api2mcp serve -openapi openapi.yaml -base https://api.internal -read-only

# streamable-HTTP
api2mcp serve -openapi openapi.yaml -base https://api.internal -read-only -http :8080

# swaggo source, whitelist a tag
api2mcp serve -swaggo docs/swagger.json -base https://api.internal -include-tag public
```

| Flag | Meaning |
|------|---------|
| `-config` | Path to a YAML config (overrides the flags below) |
| `-openapi` | OpenAPI 3 spec path |
| `-swaggo` | swaggo `swagger.json` path |
| `-base` | Upstream API base URL |
| `-http` | Serve streamable-HTTP on this addr (default: stdio) |
| `-read-only` | Expose only GET/HEAD operations |
| `-include-tag` | Whitelist a tag (repeatable) |

## Config file

For anything richer, use a YAML config:

```bash
api2mcp serve -config api2mcp.yaml
```

```yaml
name: demo-users
version: 1.0.0
baseURL: https://jsonplaceholder.typicode.com

source:
  type: openapi               # openapi | swaggo
  path: examples/openapi-stdio/openapi.yaml

transport:
  type: http                  # stdio | http
  addr: ":8080"
  path: /mcp

curation:
  readOnly: true              # safe by default: only GET/HEAD become tools
  includeTags: [public]       # whitelist
  excludePaths: ["/admin/*"]  # belt and suspenders

forwardHeaders: [Authorization]   # pass the caller's JWT upstream
maxResponseBytes: 16384           # cap so big payloads don't flood context
```

### Config reference

| Key | Type | Notes |
|-----|------|-------|
| `name`, `version` | string | Advertised to MCP clients |
| `baseURL` | string | Upstream API root |
| `source.type` | `openapi` \| `swaggo` | |
| `source.path` / `source.url` | string | Spec location |
| `transport.type` | `stdio` \| `http` | |
| `transport.addr` / `transport.path` | string | HTTP only |
| `curation.readOnly` | bool | Drop mutating operations |
| `curation.includeTags` / `excludeTags` | []string | |
| `curation.includePaths` / `excludePaths` | []string | Globs |
| `curation.includeOperations` / `excludeOperations` | []string | Exact ids |
| `forwardHeaders` | []string | Headers forwarded upstream |
| `staticHeaders` | map | Fixed headers injected on every call |
| `maxResponseBytes` | int | Response size cap |
