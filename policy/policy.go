// Package policy is the curation layer (L4): it decides which operations are
// safe and appropriate to expose as MCP tools. This is api2mcp's core value —
// not converting an API, but exposing a deliberately small, safe slice of it.
//
// The default posture is conservative: with ReadOnly enabled, only side-effect
// free operations survive, so an LLM cannot POST/PUT/DELETE against your API
// unless you explicitly opt in.
package policy

import (
	"path"
	"strings"

	"github.com/promptrails/api2mcp/ir"
)

// Matcher reports whether an operation matches some criterion.
type Matcher func(ir.Operation) bool

// Policy is an ordered set of include/exclude matchers plus a read-only switch.
//
// Filtering semantics, applied per operation:
//  1. if ReadOnly is set and the op is not read-only, drop it;
//  2. if any Include matchers exist, the op must match at least one, else drop;
//  3. if the op matches any Exclude matcher, drop it.
type Policy struct {
	Includes []Matcher
	Excludes []Matcher
	ReadOnly bool
}

// Apply returns the subset of ops permitted by the policy, preserving order.
func (p Policy) Apply(ops []ir.Operation) []ir.Operation {
	out := make([]ir.Operation, 0, len(ops))
	for _, op := range ops {
		if p.allows(op) {
			out = append(out, op)
		}
	}
	return out
}

func (p Policy) allows(op ir.Operation) bool {
	if p.ReadOnly && !op.IsReadOnly() {
		return false
	}
	if len(p.Includes) > 0 && !anyMatch(p.Includes, op) {
		return false
	}
	if anyMatch(p.Excludes, op) {
		return false
	}
	return true
}

func anyMatch(ms []Matcher, op ir.Operation) bool {
	for _, m := range ms {
		if m(op) {
			return true
		}
	}
	return false
}

// --- Matchers -------------------------------------------------------------

// Tag matches operations carrying the given tag.
func Tag(tag string) Matcher {
	return func(op ir.Operation) bool {
		for _, t := range op.Tags {
			if t == tag {
				return true
			}
		}
		return false
	}
}

// PathGlob matches operations whose path matches a shell glob (e.g. "/admin/*").
// As a convenience, a trailing "/*" matches recursively, so "/admin/*" matches
// both "/admin/users" and "/admin/users/{id}". An invalid pattern never matches.
func PathGlob(glob string) Matcher {
	if prefix, ok := strings.CutSuffix(glob, "/*"); ok {
		return func(op ir.Operation) bool {
			return op.Path == prefix || strings.HasPrefix(op.Path, prefix+"/")
		}
	}
	return func(op ir.Operation) bool {
		ok, err := path.Match(glob, op.Path)
		return err == nil && ok
	}
}

// Method matches operations using the given HTTP verb (case-insensitive).
func Method(verb string) Matcher {
	verb = strings.ToUpper(verb)
	return func(op ir.Operation) bool { return op.Method == verb }
}

// OperationID matches an operation by its exact id.
func OperationID(id string) Matcher {
	return func(op ir.Operation) bool { return op.ID == id }
}

// Custom wraps an arbitrary predicate as a Matcher.
func Custom(fn func(ir.Operation) bool) Matcher { return Matcher(fn) }
