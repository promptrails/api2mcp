package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/promptrails/api2mcp/ir"
)

func TestExecutorForwardsHeadersAndPathParams(t *testing.T) {
	var gotAuth, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	exec := &Executor{BaseURL: srv.URL}
	op := ir.Operation{ID: "getUser", Method: "GET", Path: "/users/{id}",
		Params: []ir.Param{{Name: "id", In: ir.InPath, Required: true}}}

	ctx := WithHeaders(context.Background(), map[string]string{"Authorization": "Bearer tok123"})
	resp, err := exec.Call(ctx, op, map[string]any{"id": 7})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != 200 {
		t.Fatalf("status = %d", resp.Status)
	}
	if gotAuth != "Bearer tok123" {
		t.Errorf("forwarded auth = %q, want Bearer tok123", gotAuth)
	}
	if gotPath != "/users/7" {
		t.Errorf("path = %q, want /users/7", gotPath)
	}
}

func TestExecutorSendsJSONBody(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		body = string(b)
		w.WriteHeader(201)
	}))
	defer srv.Close()

	exec := &Executor{BaseURL: srv.URL}
	op := ir.Operation{ID: "createPost", Method: "POST", Path: "/posts",
		RequestBody: &ir.Body{Required: true}}
	_, err := exec.Call(context.Background(), op, map[string]any{
		"body": map[string]any{"title": "hi"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `"title":"hi"`) {
		t.Errorf("body = %q, want it to contain title:hi", body)
	}
}

func TestTruncate(t *testing.T) {
	shape := Truncate(10, DefaultShape)
	out := shape(&Response{Status: 200, Body: []byte("0123456789ABCDEF")})
	if !strings.HasPrefix(out, "0123456789") {
		t.Errorf("missing prefix: %q", out)
	}
	if !strings.Contains(out, "truncated") {
		t.Errorf("missing truncation notice: %q", out)
	}
}
