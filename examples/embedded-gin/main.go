// Command embedded-gin shows embedded mode: an existing Gin API mounts its own
// MCP endpoint at /mcp in the same process, with no OpenAPI spec and no
// separate server. The Gin routes themselves are the source of tools.
//
//	go run ./examples/embedded-gin   # then point an MCP client at http://localhost:8080/mcp
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/promptrails/api2mcp"
	"github.com/promptrails/api2mcp/adapter/ginadapter"
)

func main() {
	r := gin.Default()

	// A couple of ordinary endpoints.
	r.GET("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, []gin.H{{"id": 1, "name": "Ada"}})
	})
	r.GET("/users/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id"), "name": "Ada"})
	})
	r.POST("/users", func(c *gin.Context) { c.JSON(http.StatusCreated, gin.H{"ok": true}) })

	// Build an MCP server from the live router; expose only read-only tools.
	srv := api2mcp.New(ginadapter.New(r),
		api2mcp.WithName("embedded-gin"),
		api2mcp.WithBaseURL("http://localhost:8080"),
		api2mcp.ReadOnly(),
		api2mcp.WithEndpointPath("/mcp"),
	)
	handler, err := srv.HTTPHandler(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Mount the MCP endpoint inside the same Gin app.
	r.Any("/mcp", gin.WrapH(handler))

	log.Println("API + MCP on http://localhost:8080 (MCP at /mcp)")
	_ = r.Run(":8080")
}
