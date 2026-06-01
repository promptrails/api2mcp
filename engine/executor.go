package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/promptrails/api2mcp/ir"
)

// Executor turns a resolved operation + arguments into a real upstream HTTP
// call. It is the only component that talks to the downstream API.
type Executor struct {
	// BaseURL is the upstream API root, e.g. "https://api.internal".
	BaseURL string
	// Client is the HTTP client used for upstream calls. Defaults to a client
	// with a 30s timeout when nil.
	Client *http.Client
	// StaticHeaders are sent on every request (e.g. an injected API key).
	StaticHeaders map[string]string
}

// Response is the raw upstream result handed back to the curation layer for
// optional shaping before it reaches the LLM.
type Response struct {
	Status      int
	ContentType string
	Body        []byte
}

func (e *Executor) client() *http.Client {
	if e.Client != nil {
		return e.Client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Call executes op with the given tool arguments and returns the upstream
// response. Path params are substituted into the URL, query/header params are
// attached, and the "body" argument (if any) is sent as a JSON body.
func (e *Executor) Call(ctx context.Context, op ir.Operation, args map[string]any) (*Response, error) {
	path := op.Path
	query := url.Values{}
	header := http.Header{}

	for _, p := range op.Params {
		val, ok := args[p.Name]
		if !ok {
			continue
		}
		s := stringify(val)
		switch p.In {
		case ir.InPath:
			path = strings.ReplaceAll(path, "{"+p.Name+"}", url.PathEscape(s))
		case ir.InQuery:
			query.Set(p.Name, s)
		case ir.InHeader:
			header.Set(p.Name, s)
		}
	}

	u := strings.TrimRight(e.BaseURL, "/") + path
	if enc := query.Encode(); enc != "" {
		u += "?" + enc
	}

	var bodyReader io.Reader
	if raw, ok := args[bodyKey]; ok && op.RequestBody != nil {
		b, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
		header.Set("Content-Type", "application/json")
	}

	req, err := http.NewRequestWithContext(ctx, op.Method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	for k, v := range e.StaticHeaders {
		req.Header.Set(k, v)
	}
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Set(k, v)
		}
	}
	// Per-call header overrides injected via context (e.g. forwarded auth).
	for k, v := range headersFromContext(ctx) {
		req.Header.Set(k, v)
	}

	resp, err := e.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("upstream call: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read upstream body: %w", err)
	}
	return &Response{
		Status:      resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Body:        data,
	}, nil
}

// stringify renders a scalar arg as a string for URL/header use.
func stringify(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}

// ctxHeaderKey is the context key under which per-call header overrides ride.
type ctxHeaderKey struct{}

// WithHeaders attaches per-call headers (e.g. a forwarded Authorization header)
// to ctx. The executor applies them last so they win over static headers.
func WithHeaders(ctx context.Context, h map[string]string) context.Context {
	return context.WithValue(ctx, ctxHeaderKey{}, h)
}

func headersFromContext(ctx context.Context) map[string]string {
	h, _ := ctx.Value(ctxHeaderKey{}).(map[string]string)
	return h
}
