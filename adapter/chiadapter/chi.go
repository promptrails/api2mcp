// Package chiadapter exposes a live chi router as an api2mcp Source.
package chiadapter

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/promptrails/api2mcp/ir"
	"github.com/promptrails/api2mcp/source/route"
)

// Source adapts a chi.Routes (e.g. *chi.Mux).
type Source struct{ router chi.Routes }

// New returns a Source backed by the given chi router.
func New(router chi.Routes) *Source { return &Source{router: router} }

// Operations implements source.Source by walking the chi route tree. chi paths
// already use the "{id}" template syntax, so no conversion is needed.
func (s *Source) Operations(_ context.Context) ([]ir.Operation, error) {
	var ops []ir.Operation
	err := chi.Walk(s.router, func(method, path string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if route.StandardMethod(method) {
			ops = append(ops, route.FromTemplate(method, path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ops, nil
}
