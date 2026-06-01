// Package fiberadapter exposes a live *fiber.App as an api2mcp Source.
package fiberadapter

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/promptrails/api2mcp/ir"
	"github.com/promptrails/api2mcp/source/route"
)

// Source adapts a *fiber.App.
type Source struct{ app *fiber.App }

// New returns a Source backed by the given Fiber app.
func New(app *fiber.App) *Source { return &Source{app: app} }

// Operations implements source.Source by walking Fiber's registered routes.
// The boolean filter drops middleware ("use") entries so only real endpoints
// become tools.
func (s *Source) Operations(_ context.Context) ([]ir.Operation, error) {
	var ops []ir.Operation
	for _, r := range s.app.GetRoutes(true) {
		if !route.StandardMethod(r.Method) {
			continue
		}
		ops = append(ops, route.FromTemplate(r.Method, route.ColonToBrace(r.Path)))
	}
	return ops, nil
}
