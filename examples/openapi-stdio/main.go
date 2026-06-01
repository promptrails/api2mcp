// Command openapi-stdio is the M0 spike: it loads an OpenAPI spec and serves
// every operation as an MCP tool over stdio.
//
//	go run ./examples/openapi-stdio -spec examples/openapi-stdio/openapi.yaml
package main

import (
	"context"
	"flag"
	"log"

	"github.com/promptrails/api2mcp"
	"github.com/promptrails/api2mcp/source/openapi"
)

func main() {
	spec := flag.String("spec", "openapi.yaml", "path to OpenAPI 3 spec")
	base := flag.String("base", "https://jsonplaceholder.typicode.com", "upstream API base URL")
	flag.Parse()

	src, err := openapi.FromFile(*spec)
	if err != nil {
		log.Fatalf("load spec: %v", err)
	}

	srv := api2mcp.New(src,
		api2mcp.WithName("demo-users"),
		api2mcp.WithBaseURL(*base),
	)

	if err := srv.ServeStdio(context.Background()); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
