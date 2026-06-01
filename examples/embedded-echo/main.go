// Command embedded-echo mounts an MCP endpoint at /mcp inside an existing Echo
// app. The Echo routes themselves are the source of tools.
//
//	go run ./examples/embedded-echo   # MCP client -> http://localhost:8770/mcp
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/promptrails/api2mcp"
	"github.com/promptrails/api2mcp/adapter/echoadapter"
)

func main() {
	e := echo.New()
	e.HideBanner = true

	e.GET("/users", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []echo.Map{{"id": 1, "name": "Ada"}})
	})
	e.GET("/users/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"id": c.Param("id"), "name": "Ada"})
	})
	e.POST("/users", func(c echo.Context) error { return c.JSON(http.StatusCreated, echo.Map{"ok": true}) })

	srv := api2mcp.New(echoadapter.New(e),
		api2mcp.WithName("embedded-echo"),
		api2mcp.WithBaseURL("http://localhost:8770"),
		api2mcp.ReadOnly(),
		api2mcp.WithEndpointPath("/mcp"),
	)
	h, err := srv.HTTPHandler(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	e.Any("/mcp", echo.WrapHandler(h))

	log.Println("API + MCP on http://localhost:8770 (MCP at /mcp)")
	_ = e.Start(":8770")
}
