package reflectschema

import (
	"context"
	"strings"
	"testing"

	"github.com/promptrails/api2mcp/ir"
	"github.com/promptrails/api2mcp/source"
)

type createUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age,omitempty"`
}

func TestEnrichAddsBodySchema(t *testing.T) {
	base := source.Static(
		ir.Operation{ID: "post_users", Method: "POST", Path: "/users"},
		ir.Operation{ID: "get_users", Method: "GET", Path: "/users"},
	)
	enr := New(base).Body("post_users", createUser{})

	ops, err := enr.Operations(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	var post, get ir.Operation
	for _, o := range ops {
		switch o.ID {
		case "post_users":
			post = o
		case "get_users":
			get = o
		}
	}

	if post.RequestBody == nil {
		t.Fatal("post_users got no request body schema")
	}
	schema := string(post.RequestBody.Schema)
	for _, want := range []string{`"name"`, `"email"`, `"age"`, `"properties"`} {
		if !strings.Contains(schema, want) {
			t.Errorf("schema missing %s: %s", want, schema)
		}
	}
	if strings.Contains(schema, "$ref") {
		t.Errorf("schema should be inlined (no $ref): %s", schema)
	}
	if get.RequestBody != nil {
		t.Error("get_users should not have been enriched")
	}
}
