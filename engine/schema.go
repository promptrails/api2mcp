package engine

import (
	"encoding/json"

	"github.com/promptrails/api2mcp/ir"
)

// InputSchema assembles a single JSON Schema object describing every input an
// LLM must provide to call the operation. The convention is deliberately flat
// and predictable:
//
//   - each path/query/header parameter becomes a top-level property keyed by
//     its name, carrying the parameter's own JSON Schema;
//   - a JSON request body becomes a single "body" property holding the body
//     schema.
//
// This keeps the mapping back to an HTTP request unambiguous in the executor.
func InputSchema(op ir.Operation) json.RawMessage {
	props := map[string]json.RawMessage{}
	var required []string

	for _, p := range op.Params {
		props[p.Name] = withDescription(p.Schema, p.Description, defaultStringSchema)
		if p.Required {
			required = append(required, p.Name)
		}
	}

	if op.RequestBody != nil {
		props[bodyKey] = withDescription(op.RequestBody.Schema, op.RequestBody.Description, defaultObjectSchema)
		if op.RequestBody.Required {
			required = append(required, bodyKey)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	out, _ := json.Marshal(schema)
	return out
}

// bodyKey is the property name under which the request body is nested.
const bodyKey = "body"

var (
	defaultStringSchema = json.RawMessage(`{"type":"string"}`)
	defaultObjectSchema = json.RawMessage(`{"type":"object"}`)
)

// withDescription returns the schema fragment with desc applied (without
// clobbering an existing description), falling back to def when raw is empty.
func withDescription(raw json.RawMessage, desc string, def json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		raw = def
	}
	if desc == "" {
		return raw
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return raw // not an object schema; leave as-is
	}
	if _, ok := m["description"]; !ok {
		m["description"] = desc
	}
	out, err := json.Marshal(m)
	if err != nil {
		return raw
	}
	return out
}
