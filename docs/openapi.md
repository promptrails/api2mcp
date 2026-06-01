# OpenAPI

The OpenAPI source is the primary, framework-agnostic path. Most Go services
already emit a `swagger.json` / `openapi.yaml`, so they can become an MCP server
without importing anything into the app.

## Loading a spec

```go
import "github.com/promptrails/api2mcp/source/openapi"

// from a local file (JSON or YAML)
src, err := openapi.FromFile("openapi.yaml")

// from a URL
src, err := openapi.FromURL("https://api.internal/openapi.json")

// from bytes
src, err := openapi.FromData(specBytes)
```

Then wire it into a server:

```go
srv := api2mcp.New(src, api2mcp.WithBaseURL("https://api.internal"))
```

## How operations map to tools

For each path + method in the spec, api2mcp generates one MCP tool:

| OpenAPI | MCP tool |
|---------|----------|
| `operationId` | tool name (synthesized from method+path if absent) |
| `summary` + `description` | tool description (always ends with `Calls GET /path`) |
| path / query / header parameters | top-level input properties, keyed by name |
| `requestBody` (application/json) | a single `body` input property |
| `tags` | used by curation filters |

The input schema is assembled as a single JSON Schema object. Path, query and
header parameters become top-level properties; a JSON request body is nested
under a `body` property. This keeps the mapping back to an HTTP request
unambiguous.

### Example

A spec operation:

```yaml
/users/{id}:
  get:
    operationId: getUser
    summary: Get a user by id
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: integer }
```

becomes a tool `getUser` with input schema:

```json
{
  "type": "object",
  "properties": { "id": { "type": "integer", "description": "..." } },
  "required": ["id"]
}
```

## Base URL

The spec's `servers` block is **not** used to route calls automatically — set
the upstream explicitly with `WithBaseURL` so you stay in control of where tool
calls actually go (prod vs staging vs a proxy):

```go
srv := api2mcp.New(src, api2mcp.WithBaseURL("https://api.internal"))
```

## Next

- [Curation & Safety](curation.md) — expose only a safe slice of the spec
- [swaggo](swaggo.md) — if your spec is OpenAPI 2.0
