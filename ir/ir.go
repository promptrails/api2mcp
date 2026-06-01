// Package ir defines the framework-agnostic intermediate representation that
// every Source normalizes into. It is intentionally a small subset of an
// OpenAPI operation: enough to build an MCP tool and execute the upstream HTTP
// call, nothing more. The whole library has exactly one core code path because
// everything — OpenAPI specs, Gin routes, Echo routes — collapses into []Operation.
package ir

import "encoding/json"

// Operation is one callable endpoint: a single HTTP method+path pair plus the
// metadata an LLM needs to call it well (names, descriptions, schemas).
type Operation struct {
	// ID is the stable, unique identifier used as the MCP tool name.
	ID string
	// Method is the upstream HTTP verb (GET, POST, ...).
	Method string
	// Path is the templated path, e.g. "/users/{id}".
	Path string

	Summary     string
	Description string
	Tags        []string

	// Params are path, query and header parameters.
	Params []Param
	// RequestBody is the (optional) request body, JSON only for now.
	RequestBody *Body

	// Security lists the names of security schemes that guard this operation.
	// Used by the curation layer (e.g. to surface which tools touch auth).
	Security []string
}

// ParamIn enumerates where a parameter is carried in the HTTP request.
type ParamIn string

const (
	InPath   ParamIn = "path"
	InQuery  ParamIn = "query"
	InHeader ParamIn = "header"
)

// Param is a single named input carried in the path, query string or headers.
type Param struct {
	Name        string
	In          ParamIn
	Required    bool
	Description string
	// Schema is a JSON Schema fragment for this parameter's value. May be nil,
	// in which case the engine defaults to {"type":"string"}.
	Schema json.RawMessage
}

// Body describes a JSON request body.
type Body struct {
	Required    bool
	Description string
	// Schema is the JSON Schema for the body object. May be nil.
	Schema json.RawMessage
}

// IsReadOnly reports whether the operation is side-effect free (safe to expose
// to an LLM by default). GET and HEAD are treated as read-only.
func (o Operation) IsReadOnly() bool {
	return o.Method == "GET" || o.Method == "HEAD"
}
