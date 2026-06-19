package source

import (
	"context"
	"errors"
	"testing"

	"github.com/promptrails/api2mcp/ir"
)

func TestStatic(t *testing.T) {
	src := Static(
		ir.Operation{ID: "a", Method: "GET", Path: "/a"},
		ir.Operation{ID: "b", Method: "POST", Path: "/b"},
	)
	ops, err := src.Operations(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 2 || ops[0].ID != "a" || ops[1].ID != "b" {
		t.Errorf("ops = %+v", ops)
	}
}

func TestFunc_AdaptsAndPropagatesError(t *testing.T) {
	wantErr := errors.New("discovery failed")
	src := Func(func(context.Context) ([]ir.Operation, error) {
		return nil, wantErr
	})
	if _, err := src.Operations(context.Background()); !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
}
