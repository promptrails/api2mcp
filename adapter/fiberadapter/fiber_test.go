package fiberadapter

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestFiberRoutesToOperations(t *testing.T) {
	app := fiber.New()
	app.Get("/users", func(*fiber.Ctx) error { return nil })
	app.Get("/users/:id", func(*fiber.Ctx) error { return nil })
	app.Post("/posts", func(*fiber.Ctx) error { return nil })

	ops, err := New(app).Operations(context.Background())
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
	if byID["post_posts"] != http.MethodPost+" /posts" {
		t.Errorf("post_posts = %q", byID["post_posts"])
	}
}
