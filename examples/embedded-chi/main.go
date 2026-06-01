// Command embedded-chi mounts an MCP endpoint at /mcp inside an existing chi
// router. The chi routes themselves are the source of tools.
//
//	go run ./examples/embedded-chi   # MCP client -> http://localhost:8772/mcp
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/promptrails/api2mcp"
	"github.com/promptrails/api2mcp/adapter/chiadapter"
)

func main() {
	r := chi.NewRouter()

	r.Get("/users", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, []map[string]any{{"id": 1, "name": "Ada"}})
	})
	r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
		writeJSON(w, map[string]any{"id": chi.URLParam(req, "id"), "name": "Ada"})
	})
	r.Post("/users", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, map[string]any{"ok": true})
	})

	// Build the source from the routes registered so far, then mount /mcp.
	srv := api2mcp.New(chiadapter.New(r),
		api2mcp.WithName("embedded-chi"),
		api2mcp.WithBaseURL("http://localhost:8772"),
		api2mcp.ReadOnly(),
		api2mcp.WithEndpointPath("/mcp"),
	)
	h, err := srv.HTTPHandler(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	r.Handle("/mcp", h)

	log.Println("API + MCP on http://localhost:8772 (MCP at /mcp)")
	_ = http.ListenAndServe(":8772", r)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
