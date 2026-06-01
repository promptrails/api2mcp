// Package api2mcp turns an existing HTTP API into a [Model Context Protocol]
// server — OpenAPI-first, framework-agnostic, and safe by default.
//
// It exposes an existing Go HTTP API to LLMs as a set of MCP tools without
// hand-writing a tool per endpoint. Point it at an OpenAPI spec (or a live
// Gin/Echo/Fiber/chi router), curate which operations are safe to expose, and
// serve over stdio or streamable-HTTP. The MCP protocol itself is handled by
// [mark3labs/mcp-go]; this package is the layer above it.
//
// # Design
//
// Everything collapses into one intermediate representation, [ir.Operation], so
// the core has exactly one code path whether the input is a swagger.json or a
// live router. The layers are:
//
//   - Sources turn some API definition into operations: [openapi], [swaggo], or
//     the framework adapters under adapter/ (ginadapter, echoadapter,
//     fiberadapter, chiadapter).
//   - The engine maps operations to MCP tools backed by a real HTTP executor.
//   - The curation layer (the options on [Server]) decides which tools are safe
//     to expose and how their responses are shaped.
//
// # Quick start
//
// Serve an OpenAPI spec over stdio (for desktop clients like Claude Desktop):
//
//	src, _ := openapi.FromFile("openapi.yaml")
//	srv := api2mcp.New(src,
//		api2mcp.WithBaseURL("https://api.internal"),
//		api2mcp.ReadOnly(),            // only GET/HEAD become tools
//		api2mcp.IncludeTags("public"), // whitelist
//	)
//	srv.ServeStdio(context.Background())
//
// Or embed an MCP endpoint inside an existing Gin app, with the live router as
// the source of tools:
//
//	src := ginadapter.New(router)
//	srv := api2mcp.New(src, api2mcp.WithBaseURL("http://localhost:8080"), api2mcp.ReadOnly())
//	h, _ := srv.HTTPHandler(context.Background())
//	router.Any("/mcp", gin.WrapH(h))
//
// # Safety
//
// api2mcp is curation-first: the hard, valuable part isn't converting an API but
// exposing only a safe slice of it. By default it offers read-only mode,
// tag/path/operation include & exclude filters, automatic MCP tool annotations
// (read-only / destructive / idempotent hints derived from the HTTP method),
// per-tool rate limiting, audit logging, response-size caps, and auth
// forwarding.
//
// Full documentation: https://promptrails.github.io/api2mcp/
//
// [Model Context Protocol]: https://modelcontextprotocol.io
// [mark3labs/mcp-go]: https://github.com/mark3labs/mcp-go
// [ir.Operation]: https://pkg.go.dev/github.com/promptrails/api2mcp/ir#Operation
// [openapi]: https://pkg.go.dev/github.com/promptrails/api2mcp/source/openapi
// [swaggo]: https://pkg.go.dev/github.com/promptrails/api2mcp/source/swaggo
package api2mcp
