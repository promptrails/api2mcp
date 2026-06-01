// Package source defines how operations are discovered. A Source is the only
// thing the engine consumes: anything that can produce []ir.Operation — an
// OpenAPI document, a live Gin router, a hand-written list — is a valid Source.
package source

import (
	"context"

	"github.com/promptrails/api2mcp/ir"
)

// Source produces the set of operations to expose as MCP tools.
type Source interface {
	Operations(ctx context.Context) ([]ir.Operation, error)
}

// Func adapts a plain function into a Source.
type Func func(ctx context.Context) ([]ir.Operation, error)

func (f Func) Operations(ctx context.Context) ([]ir.Operation, error) { return f(ctx) }

// Static returns a Source that always yields the given operations. Useful for
// tests and for manually-curated tool sets.
func Static(ops ...ir.Operation) Source {
	return Func(func(context.Context) ([]ir.Operation, error) { return ops, nil })
}
