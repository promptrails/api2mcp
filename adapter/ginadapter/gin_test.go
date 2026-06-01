package ginadapter

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGinRoutesToOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users", func(*gin.Context) {})
	r.GET("/users/:id", func(*gin.Context) {})
	r.POST("/posts", func(*gin.Context) {})

	ops, err := New(r).Operations(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]string{} // id -> "METHOD path"
	for _, o := range ops {
		got[o.ID] = o.Method + " " + o.Path
	}
	if got["get_users_id"] != "GET /users/{id}" {
		t.Errorf("get_users_id = %q", got["get_users_id"])
	}
	if got["post_posts"] != http.MethodPost+" /posts" {
		t.Errorf("post_posts = %q", got["post_posts"])
	}
	// /users/{id} must have one required path param named id.
	for _, o := range ops {
		if o.ID == "get_users_id" {
			if len(o.Params) != 1 || o.Params[0].Name != "id" || !o.Params[0].Required {
				t.Errorf("getUser params = %+v", o.Params)
			}
		}
	}
}
