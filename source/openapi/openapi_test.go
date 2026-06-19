package openapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

const specJSON = `{
  "openapi": "3.0.0",
  "info": {"title": "Test", "version": "1.0.0"},
  "components": {
    "securitySchemes": {
      "apiKey": {"type": "apiKey", "name": "X-Key", "in": "header"}
    }
  },
  "paths": {
    "/users/{id}": {
      "get": {
        "operationId": "getUser",
        "summary": "Get a user",
        "parameters": [
          {"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}
        ],
        "responses": {"200": {"description": "ok"}}
      }
    },
    "/users": {
      "post": {
        "summary": "Create a user",
        "requestBody": {
          "required": true,
          "content": {"application/json": {"schema": {"type": "object"}}}
        },
        "security": [{"apiKey": []}],
        "responses": {"200": {"description": "ok"}}
      }
    }
  }
}`

func TestFromData_Operations(t *testing.T) {
	src, err := FromData([]byte(specJSON))
	if err != nil {
		t.Fatalf("FromData: %v", err)
	}
	ops, err := src.Operations(context.Background())
	if err != nil {
		t.Fatalf("Operations: %v", err)
	}
	if len(ops) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(ops))
	}

	// Operations are emitted in sorted path order: "/users" before "/users/{id}".
	post, get := ops[0], ops[1]

	if get.ID != "getUser" || get.Method != "GET" || get.Path != "/users/{id}" {
		t.Errorf("get op = %+v", get)
	}
	if !get.IsReadOnly() {
		t.Error("GET should be read-only")
	}
	if len(get.Params) != 1 || get.Params[0].Name != "id" || !get.Params[0].Required {
		t.Errorf("get params = %+v", get.Params)
	}
	if len(get.Params[0].Schema) == 0 {
		t.Error("path param schema should be carried through")
	}

	if post.Method != "POST" || post.Path != "/users" {
		t.Errorf("post op = %+v", post)
	}
	if post.ID != "post_users" { // synthesized — no operationId in spec
		t.Errorf("post ID = %q, want post_users", post.ID)
	}
	if post.RequestBody == nil || !post.RequestBody.Required {
		t.Errorf("post body = %+v", post.RequestBody)
	}
	if len(post.Security) != 1 || post.Security[0] != "apiKey" {
		t.Errorf("post security = %v, want [apiKey]", post.Security)
	}
	if post.IsReadOnly() {
		t.Error("POST should not be read-only")
	}
}

func TestFromData_Invalid(t *testing.T) {
	if _, err := FromData([]byte("not openapi at all")); err == nil {
		t.Fatal("expected error for invalid spec data")
	}
}

func TestFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "spec.json")
	if err := os.WriteFile(path, []byte(specJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	src, err := FromFile(path)
	if err != nil {
		t.Fatalf("FromFile: %v", err)
	}
	ops, err := src.Operations(context.Background())
	if err != nil || len(ops) != 2 {
		t.Fatalf("ops=%d err=%v", len(ops), err)
	}
}

func TestFromFile_Missing(t *testing.T) {
	if _, err := FromFile(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("expected error for missing file")
	}
}
