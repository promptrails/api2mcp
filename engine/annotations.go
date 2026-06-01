package engine

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/promptrails/api2mcp/ir"
)

// annotate derives MCP tool annotations from the operation's HTTP method. These
// hints are a safety signal: MCP clients surface them to users — e.g. warning
// before invoking a destructive tool — so deriving them correctly is part of
// api2mcp's "safe by default" posture.
//
// The mapping follows HTTP semantics:
//
//	method        readOnly  destructive  idempotent
//	GET / HEAD     true        —            —          (no env modification)
//	POST           false      false        false       (additive create)
//	PUT            false      true         true         (replace, repeatable)
//	PATCH          false      true         false        (partial mutate)
//	DELETE         false      true         true         (remove, repeatable)
//
// openWorldHint is always true: a tool call reaches an external HTTP API.
func annotate(op ir.Operation) mcp.ToolAnnotation {
	a := mcp.ToolAnnotation{
		Title:         op.Summary,
		OpenWorldHint: mcp.ToBoolPtr(true),
	}
	switch op.Method {
	case "GET", "HEAD":
		a.ReadOnlyHint = mcp.ToBoolPtr(true)
		a.DestructiveHint = mcp.ToBoolPtr(false)
		a.IdempotentHint = mcp.ToBoolPtr(true)
	case "POST":
		a.ReadOnlyHint = mcp.ToBoolPtr(false)
		a.DestructiveHint = mcp.ToBoolPtr(false)
		a.IdempotentHint = mcp.ToBoolPtr(false)
	case "PUT":
		a.ReadOnlyHint = mcp.ToBoolPtr(false)
		a.DestructiveHint = mcp.ToBoolPtr(true)
		a.IdempotentHint = mcp.ToBoolPtr(true)
	case "PATCH":
		a.ReadOnlyHint = mcp.ToBoolPtr(false)
		a.DestructiveHint = mcp.ToBoolPtr(true)
		a.IdempotentHint = mcp.ToBoolPtr(false)
	case "DELETE":
		a.ReadOnlyHint = mcp.ToBoolPtr(false)
		a.DestructiveHint = mcp.ToBoolPtr(true)
		a.IdempotentHint = mcp.ToBoolPtr(true)
	}
	return a
}
