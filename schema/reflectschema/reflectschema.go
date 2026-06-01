// Package reflectschema enriches a Source's operations with request-body JSON
// Schemas reflected from Go structs. It exists for embedded mode: a router
// exposes method+path but not the request type behind each handler, so without
// help a POST tool would have an empty body schema. Register the request struct
// per operation and this fills the gap from the struct's json tags.
//
//	src := reflectschema.New(ginadapter.New(r)).
//	    Body("post_users", CreateUserRequest{})
package reflectschema

import (
	"context"
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/promptrails/api2mcp/ir"
	"github.com/promptrails/api2mcp/source"
)

// Enricher wraps a base Source and augments operations whose request struct has
// been registered by operation id.
type Enricher struct {
	base   source.Source
	bodies map[string]any
}

// New returns an Enricher around base.
func New(base source.Source) *Enricher {
	return &Enricher{base: base, bodies: map[string]any{}}
}

// Body registers the request struct for the operation with the given id. The
// struct's json tags drive the generated schema. Returns the Enricher for
// chaining.
func (e *Enricher) Body(operationID string, requestStruct any) *Enricher {
	e.bodies[operationID] = requestStruct
	return e
}

// Operations implements source.Source: it resolves the base operations and sets
// RequestBody on each one with a registered struct.
func (e *Enricher) Operations(ctx context.Context) ([]ir.Operation, error) {
	ops, err := e.base.Operations(ctx)
	if err != nil {
		return nil, err
	}
	for i := range ops {
		v, ok := e.bodies[ops[i].ID]
		if !ok {
			continue
		}
		raw, err := reflectSchema(v)
		if err != nil {
			return nil, err
		}
		ops[i].RequestBody = &ir.Body{Required: true, Schema: raw}
	}
	return ops, nil
}

// reflectSchema produces a self-contained (no $ref) JSON Schema for v.
func reflectSchema(v any) (json.RawMessage, error) {
	r := &jsonschema.Reflector{
		DoNotReference: true, // inline everything — MCP wants a standalone schema
		ExpandedStruct: true, // emit the root struct's properties at top level
		Anonymous:      true, // no generated $id
	}
	return json.Marshal(r.Reflect(v))
}
