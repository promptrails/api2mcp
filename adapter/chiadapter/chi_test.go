package chiadapter

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestChiRoutesToOperations(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/users", func(http.ResponseWriter, *http.Request) {})
	r.Get("/users/{id}", func(http.ResponseWriter, *http.Request) {})
	r.Delete("/admin/users/{id}", func(http.ResponseWriter, *http.Request) {})

	ops, err := New(r).Operations(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]string{}
	for _, o := range ops {
		byID[o.ID] = o.Method + " " + o.Path
	}
	if byID["get_users_id"] != "GET /users/{id}" {
		t.Errorf("get_users_id = %q", byID["get_users_id"])
	}
	if byID["delete_admin_users_id"] != "DELETE /admin/users/{id}" {
		t.Errorf("delete = %q", byID["delete_admin_users_id"])
	}
}
