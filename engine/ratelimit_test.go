package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/promptrails/api2mcp/ir"
)

func TestTokenBucketRefills(t *testing.T) {
	now := time.Unix(0, 0)
	b := newBucket(RateLimit{PerSecond: 1, Burst: 2})
	b.nowFn = func() time.Time { return now }
	b.last = now // pin the clock baseline to the fake time

	if !b.allow() || !b.allow() {
		t.Fatal("burst of 2 should pass")
	}
	if b.allow() {
		t.Fatal("third call should be denied")
	}
	now = now.Add(time.Second) // refill 1 token
	if !b.allow() {
		t.Fatal("after 1s a token should be available")
	}
}

func TestBuildEnforcesRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	ops := []ir.Operation{{ID: "getThing", Method: "GET", Path: "/thing"}}
	tools := Build(ops, BuildConfig{
		Executor: &Executor{BaseURL: srv.URL},
		Limit:    RateLimit{PerSecond: 0.001, Burst: 1}, // effectively one call
	})

	call := func() string {
		res, _ := tools[0].Handler(context.Background(), mcp.CallToolRequest{})
		return res.Content[0].(mcp.TextContent).Text
	}
	if got := call(); got == "rate limit exceeded for this tool; try again shortly" {
		t.Fatalf("first call should pass, got rate-limit message")
	}
	if got := call(); got != "rate limit exceeded for this tool; try again shortly" {
		t.Fatalf("second call should be rate limited, got %q", got)
	}
}

func TestAnnotatorOverridesDestructive(t *testing.T) {
	ops := []ir.Operation{{ID: "charge", Method: "POST", Path: "/charge"}}
	tools := Build(ops, BuildConfig{
		Executor: &Executor{},
		Annotator: func(op ir.Operation, base mcp.ToolAnnotation) mcp.ToolAnnotation {
			v := true
			base.DestructiveHint = &v
			return base
		},
	})
	if !deref(tools[0].MCP.Annotations.DestructiveHint) {
		t.Error("charge should be marked destructive by the annotator")
	}
}
