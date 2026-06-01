// Package openapi turns an OpenAPI 3 document into ir.Operations. This is the
// library's primary, framework-agnostic Source: most Go services already emit a
// swagger.json / openapi.yaml, so they can become an MCP server without
// importing anything into their app.
package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/promptrails/api2mcp/ir"
)

// Source loads an OpenAPI document and exposes its operations.
type Source struct {
	doc *openapi3.T
}

// FromFile loads an OpenAPI 3 spec from a local path (JSON or YAML).
func FromFile(path string) (*Source, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("load openapi %q: %w", path, err)
	}
	return &Source{doc: doc}, nil
}

// FromURL loads an OpenAPI 3 spec from a URL.
func FromURL(raw string) (*Source, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse url %q: %w", raw, err)
	}
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromURI(u)
	if err != nil {
		return nil, fmt.Errorf("load openapi %q: %w", raw, err)
	}
	return &Source{doc: doc}, nil
}

// FromData parses an OpenAPI 3 spec from raw bytes (JSON or YAML).
func FromData(data []byte) (*Source, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("parse openapi: %w", err)
	}
	return &Source{doc: doc}, nil
}

// methods is the fixed iteration order so tool generation is deterministic.
var methods = []struct {
	verb string
	get  func(*openapi3.PathItem) *openapi3.Operation
}{
	{"GET", func(p *openapi3.PathItem) *openapi3.Operation { return p.Get }},
	{"POST", func(p *openapi3.PathItem) *openapi3.Operation { return p.Post }},
	{"PUT", func(p *openapi3.PathItem) *openapi3.Operation { return p.Put }},
	{"PATCH", func(p *openapi3.PathItem) *openapi3.Operation { return p.Patch }},
	{"DELETE", func(p *openapi3.PathItem) *openapi3.Operation { return p.Delete }},
	{"HEAD", func(p *openapi3.PathItem) *openapi3.Operation { return p.Head }},
}

// Operations implements source.Source.
func (s *Source) Operations(_ context.Context) ([]ir.Operation, error) {
	var ops []ir.Operation

	paths := s.doc.Paths.Map()
	pathKeys := make([]string, 0, len(paths))
	for k := range paths {
		pathKeys = append(pathKeys, k)
	}
	sort.Strings(pathKeys)

	for _, path := range pathKeys {
		item := paths[path]
		for _, m := range methods {
			op := m.get(item)
			if op == nil {
				continue
			}
			ops = append(ops, convert(m.verb, path, item, op))
		}
	}
	return ops, nil
}

func convert(verb, path string, item *openapi3.PathItem, op *openapi3.Operation) ir.Operation {
	out := ir.Operation{
		ID:          operationID(op, verb, path),
		Method:      verb,
		Path:        path,
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
	}

	// Path-item parameters apply to every operation under it; merge them first
	// so operation-level params can override by (name,in).
	for _, ref := range append(append(openapi3.Parameters{}, item.Parameters...), op.Parameters...) {
		if ref.Value == nil {
			continue
		}
		out.Params = append(out.Params, convertParam(ref.Value))
	}

	if op.RequestBody != nil && op.RequestBody.Value != nil {
		if body := convertBody(op.RequestBody.Value); body != nil {
			out.RequestBody = body
		}
	}

	out.Security = securityNames(op)

	return out
}

func convertParam(p *openapi3.Parameter) ir.Param {
	param := ir.Param{
		Name:        p.Name,
		In:          ir.ParamIn(p.In),
		Required:    p.Required,
		Description: p.Description,
	}
	if p.Schema != nil && p.Schema.Value != nil {
		if raw, err := p.Schema.Value.MarshalJSON(); err == nil {
			param.Schema = raw
		}
	}
	return param
}

func convertBody(b *openapi3.RequestBody) *ir.Body {
	mt := b.Content.Get("application/json")
	if mt == nil || mt.Schema == nil || mt.Schema.Value == nil {
		return nil
	}
	body := &ir.Body{Required: b.Required, Description: b.Description}
	if raw, err := mt.Schema.Value.MarshalJSON(); err == nil {
		body.Schema = raw
	}
	return body
}

// operationID returns op.OperationID, or synthesizes a stable, readable one
// from the verb and path when the spec omits it.
func operationID(op *openapi3.Operation, verb, path string) string {
	if op.OperationID != "" {
		return sanitize(op.OperationID)
	}
	parts := []string{strings.ToLower(verb)}
	for _, seg := range strings.Split(path, "/") {
		if seg == "" {
			continue
		}
		seg = strings.Trim(seg, "{}")
		parts = append(parts, seg)
	}
	return sanitize(strings.Join(parts, "_"))
}

// sanitize keeps tool names to [A-Za-z0-9_], which MCP clients handle reliably.
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

func securityNames(op *openapi3.Operation) []string {
	if op.Security == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var names []string
	for _, req := range *op.Security {
		for name := range req {
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

var _ = json.Marshal
