package api2mcp

import (
	"context"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
	"github.com/promptrails/api2mcp/engine"
)

// ServeHTTP builds the server and serves it over the MCP streamable-HTTP
// transport at addr (e.g. ":8080"). This is the load-balancer-friendly
// transport suitable for remote/hosted deployments. Forwarded auth headers
// (see WithForwardHeaders) are passed through to upstream calls per request.
func (s *Server) ServeHTTP(ctx context.Context, addr string) error {
	mcpSrv, err := s.MCPServer(ctx)
	if err != nil {
		return err
	}
	opts := []server.StreamableHTTPOption{
		server.WithStateLess(true),
	}
	if s.opts.endpointPath != "" {
		opts = append(opts, server.WithEndpointPath(s.opts.endpointPath))
	}
	if len(s.opts.forwardHeaders) > 0 {
		opts = append(opts, server.WithHTTPContextFunc(s.forwardHeaderFunc()))
	}
	httpSrv := server.NewStreamableHTTPServer(mcpSrv, opts...)
	return httpSrv.Start(addr)
}

// HTTPHandler returns the streamable-HTTP handler so callers can mount the MCP
// endpoint inside their own http.Server / router instead of calling ServeHTTP.
func (s *Server) HTTPHandler(ctx context.Context) (http.Handler, error) {
	mcpSrv, err := s.MCPServer(ctx)
	if err != nil {
		return nil, err
	}
	opts := []server.StreamableHTTPOption{server.WithStateLess(true)}
	if s.opts.endpointPath != "" {
		opts = append(opts, server.WithEndpointPath(s.opts.endpointPath))
	}
	if len(s.opts.forwardHeaders) > 0 {
		opts = append(opts, server.WithHTTPContextFunc(s.forwardHeaderFunc()))
	}
	return server.NewStreamableHTTPServer(mcpSrv, opts...), nil
}

// forwardHeaderFunc copies the configured request headers from the incoming MCP
// HTTP request into the context, so the executor re-sends them upstream. This
// is how a caller's Authorization (Bearer/JWT) reaches the protected API.
func (s *Server) forwardHeaderFunc() server.HTTPContextFunc {
	names := s.opts.forwardHeaders
	return func(ctx context.Context, r *http.Request) context.Context {
		fwd := make(map[string]string, len(names))
		for _, name := range names {
			if v := r.Header.Get(name); v != "" {
				fwd[name] = v
			}
		}
		if len(fwd) == 0 {
			return ctx
		}
		return engine.WithHeaders(ctx, fwd)
	}
}
