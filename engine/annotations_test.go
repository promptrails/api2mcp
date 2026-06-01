package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/promptrails/api2mcp/ir"
)

func deref(b *bool) bool { return b != nil && *b }

func TestAnnotateByMethod(t *testing.T) {
	cases := []struct {
		method                      string
		readOnly, destr, idempotent bool
	}{
		{"GET", true, false, true},
		{"HEAD", true, false, true},
		{"POST", false, false, false},
		{"PUT", false, true, true},
		{"PATCH", false, true, false},
		{"DELETE", false, true, true},
	}
	for _, c := range cases {
		a := annotate(ir.Operation{Method: c.method})
		if deref(a.ReadOnlyHint) != c.readOnly || deref(a.DestructiveHint) != c.destr || deref(a.IdempotentHint) != c.idempotent {
			t.Errorf("%s: got ro=%v destr=%v idem=%v", c.method,
				deref(a.ReadOnlyHint), deref(a.DestructiveHint), deref(a.IdempotentHint))
		}
		if !deref(a.OpenWorldHint) {
			t.Errorf("%s: openWorldHint should be true", c.method)
		}
	}
}

func TestBuildSetsAnnotationsAndAudits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	var events []AuditEvent
	ops := []ir.Operation{{ID: "createUser", Method: "POST", Path: "/users", RequestBody: &ir.Body{}}}
	tools := Build(ops, BuildConfig{
		Executor: &Executor{BaseURL: srv.URL},
		Audit:    func(e AuditEvent) { events = append(events, e) },
	})

	if len(tools) != 1 {
		t.Fatalf("got %d tools", len(tools))
	}
	if deref(tools[0].MCP.Annotations.ReadOnlyHint) {
		t.Error("POST tool should not be read-only")
	}

	_, err := tools[0].Handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(events))
	}
	if events[0].OperationID != "createUser" || events[0].Status != 201 {
		t.Errorf("audit event = %+v", events[0])
	}
}
