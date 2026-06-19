package swaggo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// A minimal swaggo-style Swagger (OpenAPI 2.0) document.
const swagger2JSON = `{
  "swagger": "2.0",
  "info": {"title": "Test", "version": "1.0.0"},
  "paths": {
    "/ping": {
      "get": {
        "operationId": "ping",
        "summary": "Health check",
        "responses": {"200": {"description": "ok"}}
      }
    }
  }
}`

func TestFromData_ConvertsToOperations(t *testing.T) {
	src, err := FromData([]byte(swagger2JSON))
	if err != nil {
		t.Fatalf("FromData: %v", err)
	}
	ops, err := src.Operations(context.Background())
	if err != nil {
		t.Fatalf("Operations: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	if ops[0].Method != "GET" || ops[0].Path != "/ping" {
		t.Errorf("op = %+v", ops[0])
	}
	if ops[0].ID != "ping" {
		t.Errorf("ID = %q, want ping", ops[0].ID)
	}
}

func TestFromData_Invalid(t *testing.T) {
	if _, err := FromData([]byte("{not swagger")); err == nil {
		t.Fatal("expected error for invalid swagger data")
	}
}

func TestFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "swagger.json")
	if err := os.WriteFile(path, []byte(swagger2JSON), 0o600); err != nil {
		t.Fatal(err)
	}
	src, err := FromFile(path)
	if err != nil {
		t.Fatalf("FromFile: %v", err)
	}
	ops, err := src.Operations(context.Background())
	if err != nil || len(ops) != 1 {
		t.Fatalf("ops=%d err=%v", len(ops), err)
	}
}
