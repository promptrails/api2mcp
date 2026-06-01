// Package api2mcp turns an existing HTTP API into an MCP server.
//
// It is OpenAPI-first and layered: a Source (an OpenAPI spec, a live router, a
// manual list) is normalized into a framework-agnostic intermediate
// representation, the engine maps that to MCP tools backed by real HTTP calls,
// and a curation layer decides which tools are safe to expose and how their
// responses are shaped. The MCP protocol itself is handled by mark3labs/mcp-go.
//
// Minimal use:
//
//	src, _ := openapi.FromFile("openapi.yaml")
//	srv := api2mcp.New(src, api2mcp.WithBaseURL("https://api.internal"))
//	srv.ServeStdio(context.Background())
package api2mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
	"github.com/promptrails/api2mcp/engine"
	"github.com/promptrails/api2mcp/ir"
	"github.com/promptrails/api2mcp/policy"
	"github.com/promptrails/api2mcp/source"
)

// Server wires a Source to the engine and an MCP transport.
type Server struct {
	src  source.Source
	opts options
}

type options struct {
	name          string
	version       string
	baseURL       string
	httpClient    *http.Client
	staticHeaders map[string]string
	shape         engine.ShapeFunc
	policy        policy.Policy
}

// Option configures a Server.
type Option func(*options)

// WithBaseURL sets the upstream API root that tool calls are sent to. Required
// unless the Source already resolves absolute URLs.
func WithBaseURL(url string) Option { return func(o *options) { o.baseURL = url } }

// WithName sets the MCP server name advertised to clients.
func WithName(name string) Option { return func(o *options) { o.name = name } }

// WithVersion sets the MCP server version advertised to clients.
func WithVersion(v string) Option { return func(o *options) { o.version = v } }

// WithHTTPClient overrides the HTTP client used for upstream calls.
func WithHTTPClient(c *http.Client) Option { return func(o *options) { o.httpClient = c } }

// WithStaticHeader adds a header sent on every upstream call (e.g. an API key).
func WithStaticHeader(key, value string) Option {
	return func(o *options) {
		if o.staticHeaders == nil {
			o.staticHeaders = map[string]string{}
		}
		o.staticHeaders[key] = value
	}
}

// WithResponseShaper overrides how upstream responses are rendered for the LLM.
func WithResponseShaper(s engine.ShapeFunc) Option { return func(o *options) { o.shape = s } }

// --- Curation (L4) --------------------------------------------------------

// ReadOnly exposes only side-effect free operations (GET/HEAD). This is the
// recommended default when pointing an LLM at a real API.
func ReadOnly() Option { return func(o *options) { o.policy.ReadOnly = true } }

// IncludeTags keeps only operations carrying at least one of the given tags
// (a whitelist). Combine with other Include* options — they are OR-ed.
func IncludeTags(tags ...string) Option {
	return func(o *options) {
		for _, t := range tags {
			o.policy.Includes = append(o.policy.Includes, policy.Tag(t))
		}
	}
}

// ExcludeTags drops operations carrying any of the given tags.
func ExcludeTags(tags ...string) Option {
	return func(o *options) {
		for _, t := range tags {
			o.policy.Excludes = append(o.policy.Excludes, policy.Tag(t))
		}
	}
}

// IncludePaths keeps only operations whose path matches one of the globs
// (e.g. "/public/*").
func IncludePaths(globs ...string) Option {
	return func(o *options) {
		for _, g := range globs {
			o.policy.Includes = append(o.policy.Includes, policy.PathGlob(g))
		}
	}
}

// ExcludePaths drops operations whose path matches one of the globs
// (e.g. "/admin/*", "/internal/*").
func ExcludePaths(globs ...string) Option {
	return func(o *options) {
		for _, g := range globs {
			o.policy.Excludes = append(o.policy.Excludes, policy.PathGlob(g))
		}
	}
}

// IncludeOperations keeps only the operations with the given ids.
func IncludeOperations(ids ...string) Option {
	return func(o *options) {
		for _, id := range ids {
			o.policy.Includes = append(o.policy.Includes, policy.OperationID(id))
		}
	}
}

// ExcludeOperations drops the operations with the given ids.
func ExcludeOperations(ids ...string) Option {
	return func(o *options) {
		for _, id := range ids {
			o.policy.Excludes = append(o.policy.Excludes, policy.OperationID(id))
		}
	}
}

// WithFilter adds a custom predicate; operations for which it returns false are
// dropped. (Implemented as an exclude of the negation.)
func WithFilter(keep func(ir.Operation) bool) Option {
	return func(o *options) {
		o.policy.Excludes = append(o.policy.Excludes, policy.Custom(func(op ir.Operation) bool {
			return !keep(op)
		}))
	}
}

// curate applies the curation policy before tool building.
func (o options) curate(ops []ir.Operation) []ir.Operation { return o.policy.Apply(ops) }

// New creates a Server from a Source and options.
func New(src source.Source, opts ...Option) *Server {
	o := options{name: "api2mcp", version: "0.1.0"}
	for _, fn := range opts {
		fn(&o)
	}
	return &Server{src: src, opts: o}
}

// executor builds the HTTP executor from options.
func (s *Server) executor() *engine.Executor {
	return &engine.Executor{
		BaseURL:       s.opts.baseURL,
		Client:        s.opts.httpClient,
		StaticHeaders: s.opts.staticHeaders,
	}
}

// Tools resolves the Source and builds the MCP tools (after curation).
func (s *Server) Tools(ctx context.Context) ([]engine.Tool, error) {
	ops, err := s.src.Operations(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve operations: %w", err)
	}
	ops = s.opts.curate(ops)
	return engine.Build(ops, s.executor(), s.opts.shape), nil
}

// MCPServer builds a ready-to-serve mcp-go server with all tools registered.
func (s *Server) MCPServer(ctx context.Context) (*server.MCPServer, error) {
	tools, err := s.Tools(ctx)
	if err != nil {
		return nil, err
	}
	mcpSrv := server.NewMCPServer(s.opts.name, s.opts.version, server.WithToolCapabilities(true))
	engine.Register(mcpSrv, tools)
	return mcpSrv, nil
}

// ServeStdio builds the server and serves it over stdio (for desktop clients
// like Claude Desktop, Cursor, Zed).
func (s *Server) ServeStdio(ctx context.Context) error {
	mcpSrv, err := s.MCPServer(ctx)
	if err != nil {
		return err
	}
	return server.ServeStdio(mcpSrv)
}
