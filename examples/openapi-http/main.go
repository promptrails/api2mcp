// Command openapi-http serves an OpenAPI spec as an MCP server over the
// streamable-HTTP transport, forwarding the caller's Authorization header to
// the upstream API and capping response size.
//
//	go run ./examples/openapi-http -spec examples/openapi-stdio/openapi.yaml -addr :8080
package main

import (
	"context"
	"flag"
	"log"

	"github.com/promptrails/api2mcp"
	"github.com/promptrails/api2mcp/source/openapi"
)

func main() {
	spec := flag.String("spec", "examples/openapi-stdio/openapi.yaml", "OpenAPI 3 spec")
	base := flag.String("base", "https://jsonplaceholder.typicode.com", "upstream base URL")
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	src, err := openapi.FromFile(*spec)
	if err != nil {
		log.Fatalf("load spec: %v", err)
	}

	srv := api2mcp.New(src,
		api2mcp.WithName("demo-users"),
		api2mcp.WithBaseURL(*base),
		api2mcp.ReadOnly(),
		api2mcp.IncludeTags("public"),
		api2mcp.WithForwardHeaders("Authorization"), // pass client JWT upstream
		api2mcp.WithMaxResponseBytes(8<<10),          // 8 KiB cap
		api2mcp.WithEndpointPath("/mcp"),
	)

	log.Printf("MCP streamable-HTTP listening on %s/mcp", *addr)
	if err := srv.ServeHTTP(context.Background(), *addr); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
