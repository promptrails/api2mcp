// Package ginadapter exposes a live *gin.Engine as an api2mcp Source, for
// embedded mode: mount an MCP endpoint inside an existing Gin app without a
// separate process or an OpenAPI spec.
//
//	r := gin.Default()
//	// ... your routes ...
//	src := ginadapter.New(r)
//	srv := api2mcp.New(src, api2mcp.WithBaseURL("http://localhost:8080"), api2mcp.ReadOnly())
package ginadapter

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/promptrails/api2mcp/ir"
	"github.com/promptrails/api2mcp/source/route"
)

// Source adapts a *gin.Engine.
type Source struct{ engine *gin.Engine }

// New returns a Source backed by the given Gin engine.
func New(engine *gin.Engine) *Source { return &Source{engine: engine} }

// Operations implements source.Source by walking the engine's registered routes.
func (s *Source) Operations(_ context.Context) ([]ir.Operation, error) {
	var ops []ir.Operation
	for _, ri := range s.engine.Routes() {
		if !route.StandardMethod(ri.Method) {
			continue
		}
		ops = append(ops, route.FromTemplate(ri.Method, route.ColonToBrace(ri.Path)))
	}
	return ops, nil
}
