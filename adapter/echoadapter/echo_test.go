package echoadapter

import (
	"context"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestEchoRoutesToOperations(t *testing.T) {
	e := echo.New()
	e.GET("/users", func(echo.Context) error { return nil })
	e.GET("/users/:id", func(echo.Context) error { return nil })
	e.POST("/posts", func(echo.Context) error { return nil })

	ops, err := New(e).Operations(context.Background())
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
	for _, o := range ops {
		if o.ID == "get_users_id" {
			if len(o.Params) != 1 || o.Params[0].Name != "id" {
				t.Errorf("params = %+v", o.Params)
			}
		}
	}
}
