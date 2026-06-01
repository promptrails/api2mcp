// Package route holds the shared logic every framework adapter uses to turn a
// (method, path) route into an ir.Operation. Adapters differ only in how they
// enumerate routes and in their path-parameter syntax; everything downstream of
// that is identical, so it lives here once.
package route

import (
	"strings"

	"github.com/promptrails/api2mcp/ir"
)

// FromTemplate builds an ir.Operation from an HTTP method and an OpenAPI-style
// templated path such as "/users/{id}". Each "{name}" segment becomes a
// required string path parameter. Routers carry no body/query schema, so those
// are left empty for a schema provider (M4) to enrich later.
func FromTemplate(method, tmpl string) ir.Operation {
	method = strings.ToUpper(method)
	op := ir.Operation{
		ID:     synthID(method, tmpl),
		Method: method,
		Path:   tmpl,
	}
	for _, name := range pathParams(tmpl) {
		op.Params = append(op.Params, ir.Param{
			Name:     name,
			In:       ir.InPath,
			Required: true,
		})
	}
	return op
}

// ColonToBrace converts router-style path params (":id", "*filepath") into the
// OpenAPI "{id}" form used by the IR. gin, echo and fiber all use the colon/star
// syntax; chi already uses braces and skips this.
func ColonToBrace(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		switch {
		case strings.HasPrefix(s, ":"):
			segs[i] = "{" + s[1:] + "}"
		case strings.HasPrefix(s, "*"):
			name := s[1:]
			if name == "" {
				name = "wildcard"
			}
			segs[i] = "{" + name + "}"
		}
	}
	return strings.Join(segs, "/")
}

// pathParams returns the names inside "{...}" segments, in order.
func pathParams(tmpl string) []string {
	var out []string
	for _, s := range strings.Split(tmpl, "/") {
		if len(s) >= 2 && s[0] == '{' && s[len(s)-1] == '}' {
			out = append(out, s[1:len(s)-1])
		}
	}
	return out
}

// synthID builds a stable, readable tool name from the verb and path.
func synthID(method, tmpl string) string {
	parts := []string{strings.ToLower(method)}
	for _, seg := range strings.Split(tmpl, "/") {
		if seg == "" {
			continue
		}
		seg = strings.Trim(seg, "{}")
		parts = append(parts, seg)
	}
	return sanitize(strings.Join(parts, "_"))
}

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

// StandardMethod reports whether m is an HTTP verb worth exposing as a tool
// (filters out router-internal entries and the rarely-useful TRACE/CONNECT).
func StandardMethod(m string) bool {
	switch strings.ToUpper(m) {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD":
		return true
	}
	return false
}
