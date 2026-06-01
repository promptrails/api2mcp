// Package echoadapter exposes a live *echo.Echo as an api2mcp Source.
package echoadapter

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/promptrails/api2mcp/ir"
	"github.com/promptrails/api2mcp/source/route"
)

// Source adapts an *echo.Echo.
type Source struct{ e *echo.Echo }

// New returns a Source backed by the given Echo instance.
func New(e *echo.Echo) *Source { return &Source{e: e} }

// Operations implements source.Source by walking Echo's registered routes.
func (s *Source) Operations(_ context.Context) ([]ir.Operation, error) {
	var ops []ir.Operation
	for _, r := range s.e.Routes() {
		if !route.StandardMethod(r.Method) {
			continue
		}
		ops = append(ops, route.FromTemplate(r.Method, route.ColonToBrace(r.Path)))
	}
	return ops, nil
}
