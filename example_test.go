package api2mcp_test

import (
	"context"
	"fmt"

	"github.com/promptrails/api2mcp"
	"github.com/promptrails/api2mcp/ir"
	"github.com/promptrails/api2mcp/source"
)

// Expose a curated set of operations as MCP tools. ReadOnly drops write
// operations, so the DELETE is filtered out and only the GET survives.
func ExampleNew() {
	src := source.Static(
		ir.Operation{ID: "list_users", Method: "GET", Path: "/users"},
		ir.Operation{ID: "delete_user", Method: "DELETE", Path: "/users/{id}"},
	)

	srv := api2mcp.New(src,
		api2mcp.WithBaseURL("https://api.example.com"),
		api2mcp.ReadOnly(),
	)

	tools, err := srv.Tools(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(len(tools))
	fmt.Println(tools[0].Operation.ID)
	// Output:
	// 1
	// list_users
}

// Curation can also exclude specific operations by ID.
func ExampleExcludeOperations() {
	src := source.Static(
		ir.Operation{ID: "a", Method: "GET", Path: "/a"},
		ir.Operation{ID: "b", Method: "GET", Path: "/b"},
	)

	srv := api2mcp.New(src,
		api2mcp.WithBaseURL("https://api.example.com"),
		api2mcp.ExcludeOperations("b"),
	)

	tools, _ := srv.Tools(context.Background())
	for _, tl := range tools {
		fmt.Println(tl.Operation.ID)
	}
	// Output:
	// a
}
