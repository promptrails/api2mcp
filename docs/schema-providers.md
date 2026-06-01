# Schema Providers

In embedded mode a router exposes method + path but not the Go types behind each
handler, so a `POST` tool would have an empty body schema. Schema providers fill
that gap.

## Struct reflection

`schema/reflectschema` is a Source decorator: register a request struct per
operation id and it reflects the struct's json tags into a request-body JSON
Schema.

```go
import (
    "github.com/promptrails/api2mcp/adapter/ginadapter"
    "github.com/promptrails/api2mcp/schema/reflectschema"
)

type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age,omitempty"`
}

src := reflectschema.New(ginadapter.New(router)).
    Body("post_users", CreateUserRequest{})

srv := api2mcp.New(src, api2mcp.WithBaseURL("http://localhost:8080"))
```

Now the `post_users` tool advertises a full body schema:

```json
{
  "type": "object",
  "properties": {
    "name":  { "type": "string" },
    "email": { "type": "string" },
    "age":   { "type": "integer" }
  },
  "required": ["name", "email"]
}
```

The generated schema is **inlined** (no `$ref`), because MCP clients expect a
self-contained input schema per tool.

## Finding the operation id

`Body()` keys on the operation id — the same id used as the tool name. For
adapters this is the synthesized `method_path` form (e.g. `post_users`,
`get_users_id`). List your tools once (e.g. via `tools/list`) to confirm the
ids, then register structs against them.

## When you don't need this

- Using an **OpenAPI** or **swaggo** source — schemas already come from the
  spec; reflection is unnecessary.
- Exposing **read-only** tools only — GETs rarely have a request body.

Reflection is specifically for *embedded mode + writes*.
