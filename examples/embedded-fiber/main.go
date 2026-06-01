// Command embedded-fiber mounts an MCP endpoint at /mcp inside an existing
// Fiber app. The Fiber routes themselves are the source of tools.
//
//	go run ./examples/embedded-fiber   # MCP client -> http://localhost:8771/mcp
package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/promptrails/api2mcp"
	"github.com/promptrails/api2mcp/adapter/fiberadapter"
)

func main() {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Get("/users", func(c *fiber.Ctx) error {
		return c.JSON([]fiber.Map{{"id": 1, "name": "Ada"}})
	})
	app.Get("/users/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"id": c.Params("id"), "name": "Ada"})
	})
	app.Post("/users", func(c *fiber.Ctx) error { return c.Status(201).JSON(fiber.Map{"ok": true}) })

	srv := api2mcp.New(fiberadapter.New(app),
		api2mcp.WithName("embedded-fiber"),
		api2mcp.WithBaseURL("http://localhost:8771"),
		api2mcp.ReadOnly(),
		api2mcp.WithEndpointPath("/mcp"),
	)
	h, err := srv.HTTPHandler(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	app.All("/mcp", adaptor.HTTPHandler(h))

	log.Println("API + MCP on http://localhost:8771 (MCP at /mcp)")
	_ = app.Listen(":8771")
}
