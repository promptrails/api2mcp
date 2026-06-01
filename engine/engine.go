// Package engine is the framework-agnostic core: it turns ir.Operations into
// MCP tools (with raw JSON Schema inputs) and wires each tool's handler to the
// HTTP Executor. It knows nothing about OpenAPI, Gin, Echo or any source.
package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/promptrails/api2mcp/ir"
)

// AuditEvent records one tool invocation for observability. Logging every call
// — what was invoked, against which upstream path, with what result — is part
// of running an LLM-facing API safely.
type AuditEvent struct {
	OperationID string
	Method      string
	Path        string
	Status      int
	Duration    time.Duration
	Err         error
}

// AuditFunc receives an AuditEvent after each tool call completes.
type AuditFunc func(AuditEvent)

// Tool is a built MCP tool paired with its handler, ready to register.
type Tool struct {
	Operation ir.Operation
	MCP       mcp.Tool
	Handler   server.ToolHandlerFunc
}

// ShapeFunc renders an upstream response into the text returned to the LLM. It
// is the hook the curation layer uses to truncate or field-select large
// payloads. When nil, DefaultShape is used.
type ShapeFunc func(*Response) string

// AnnotateFunc lets callers adjust a tool's safety annotations after they are
// derived from the HTTP method — e.g. to force destructiveHint on a POST that
// charges a card. base is the method-derived annotation.
type AnnotateFunc func(op ir.Operation, base mcp.ToolAnnotation) mcp.ToolAnnotation

// BuildConfig holds everything Build needs beyond the operations themselves.
// All fields are optional except Executor.
type BuildConfig struct {
	Executor  *Executor
	Shape     ShapeFunc    // nil → DefaultShape
	Audit     AuditFunc    // nil → no audit logging
	Limit     RateLimit    // zero → unlimited
	Annotator AnnotateFunc // nil → method-derived annotations only
}

// Build constructs one Tool per operation.
func Build(ops []ir.Operation, cfg BuildConfig) []Tool {
	shape := cfg.Shape
	if shape == nil {
		shape = DefaultShape
	}
	tools := make([]Tool, 0, len(ops))
	for _, op := range ops {
		op := op
		t := mcp.NewToolWithRawSchema(op.ID, describe(op), InputSchema(op))
		ann := annotate(op) // safety hints derived from the HTTP method
		if cfg.Annotator != nil {
			ann = cfg.Annotator(op, ann)
		}
		t.Annotations = ann

		var bucket *tokenBucket
		if cfg.Limit.enabled() {
			bucket = newBucket(cfg.Limit)
		}

		handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if bucket != nil && !bucket.allow() {
				return mcp.NewToolResultError("rate limit exceeded for this tool; try again shortly"), nil
			}
			start := time.Now()
			resp, err := cfg.Executor.Call(ctx, op, req.GetArguments())
			if cfg.Audit != nil {
				ev := AuditEvent{OperationID: op.ID, Method: op.Method, Path: op.Path, Duration: time.Since(start), Err: err}
				if resp != nil {
					ev.Status = resp.Status
				}
				cfg.Audit(ev)
			}
			if err != nil {
				return mcp.NewToolResultErrorFromErr("upstream call failed", err), nil
			}
			text := shape(resp)
			if resp.Status >= 400 {
				return mcp.NewToolResultError(text), nil
			}
			return mcp.NewToolResultText(text), nil
		}
		tools = append(tools, Tool{Operation: op, MCP: t, Handler: handler})
	}
	return tools
}

// Register adds every built tool to an mcp-go server.
func Register(s *server.MCPServer, tools []Tool) {
	for _, t := range tools {
		s.AddTool(t.MCP, t.Handler)
	}
}

// describe builds the human/LLM-facing tool description from the operation's
// metadata, always ending with the concrete method+path for grounding.
func describe(op ir.Operation) string {
	var b strings.Builder
	if op.Summary != "" {
		b.WriteString(op.Summary)
	}
	if op.Description != "" {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(op.Description)
	}
	if b.Len() > 0 {
		b.WriteString("\n\n")
	}
	fmt.Fprintf(&b, "Calls %s %s", op.Method, op.Path)
	return b.String()
}

// DefaultShape returns the body verbatim, prefixed with the HTTP status when
// it is not a 2xx so the LLM can reason about failures.
func DefaultShape(r *Response) string {
	body := string(r.Body)
	if r.Status >= 200 && r.Status < 300 {
		return body
	}
	return fmt.Sprintf("HTTP %d\n%s", r.Status, body)
}
