package ir

import "testing"

func TestOperation_IsReadOnly(t *testing.T) {
	tests := []struct {
		method string
		want   bool
	}{
		{"GET", true},
		{"HEAD", true},
		{"POST", false},
		{"PUT", false},
		{"PATCH", false},
		{"DELETE", false},
	}
	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			op := Operation{Method: tc.method}
			if got := op.IsReadOnly(); got != tc.want {
				t.Errorf("IsReadOnly(%s) = %v, want %v", tc.method, got, tc.want)
			}
		})
	}
}
