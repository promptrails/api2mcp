package route

import (
	"testing"

	"github.com/promptrails/api2mcp/ir"
)

func TestFromTemplate(t *testing.T) {
	op := FromTemplate("get", "/users/{id}/posts/{postID}")

	if op.Method != "GET" {
		t.Errorf("Method = %q, want GET (uppercased)", op.Method)
	}
	if op.Path != "/users/{id}/posts/{postID}" {
		t.Errorf("Path = %q", op.Path)
	}
	if op.ID != "get_users_id_posts_postID" {
		t.Errorf("ID = %q, want get_users_id_posts_postID", op.ID)
	}
	if len(op.Params) != 2 {
		t.Fatalf("expected 2 path params, got %d", len(op.Params))
	}
	for _, p := range op.Params {
		if p.In != ir.InPath || !p.Required {
			t.Errorf("param %q should be a required path param: %+v", p.Name, p)
		}
	}
	if op.Params[0].Name != "id" || op.Params[1].Name != "postID" {
		t.Errorf("param names = %q, %q", op.Params[0].Name, op.Params[1].Name)
	}
}

func TestFromTemplate_NoParams(t *testing.T) {
	op := FromTemplate("POST", "/health")
	if len(op.Params) != 0 {
		t.Errorf("expected no params, got %d", len(op.Params))
	}
	if op.ID != "post_health" {
		t.Errorf("ID = %q, want post_health", op.ID)
	}
}

func TestFromTemplate_SanitizesID(t *testing.T) {
	op := FromTemplate("GET", "/v1/users.json")
	if op.ID != "get_v1_users_json" {
		t.Errorf("ID = %q, want get_v1_users_json (dot sanitized)", op.ID)
	}
}

func TestColonToBrace(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"/users/:id", "/users/{id}"},
		{"/users/:id/posts/:postID", "/users/{id}/posts/{postID}"},
		{"/files/*filepath", "/files/{filepath}"},
		{"/x/*", "/x/{wildcard}"},     // anonymous wildcard gets a name
		{"/a/{b}", "/a/{b}"},          // chi-style braces left untouched
		{"/static", "/static"},        // no params
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := ColonToBrace(tc.in); got != tc.want {
				t.Errorf("ColonToBrace(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestStandardMethod(t *testing.T) {
	for _, m := range []string{"GET", "post", "Put", "PATCH", "delete", "HEAD"} {
		if !StandardMethod(m) {
			t.Errorf("StandardMethod(%q) = false, want true", m)
		}
	}
	for _, m := range []string{"OPTIONS", "TRACE", "CONNECT", "WEIRD", ""} {
		if StandardMethod(m) {
			t.Errorf("StandardMethod(%q) = true, want false", m)
		}
	}
}
