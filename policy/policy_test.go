package policy

import (
	"testing"

	"github.com/promptrails/api2mcp/ir"
)

func ops() []ir.Operation {
	return []ir.Operation{
		{ID: "listUsers", Method: "GET", Path: "/users", Tags: []string{"users", "public"}},
		{ID: "getUser", Method: "GET", Path: "/users/{id}", Tags: []string{"users", "public"}},
		{ID: "createPost", Method: "POST", Path: "/posts", Tags: []string{"posts", "write"}},
		{ID: "deleteUser", Method: "DELETE", Path: "/admin/users/{id}", Tags: []string{"admin"}},
	}
}

func ids(ops []ir.Operation) []string {
	out := make([]string, len(ops))
	for i, o := range ops {
		out[i] = o.ID
	}
	return out
}

func TestReadOnlyDropsMutations(t *testing.T) {
	got := ids(Policy{ReadOnly: true}.Apply(ops()))
	want := []string{"listUsers", "getUser"}
	assertEqual(t, got, want)
}

func TestIncludeTagWhitelist(t *testing.T) {
	got := ids(Policy{Includes: []Matcher{Tag("public")}}.Apply(ops()))
	assertEqual(t, got, []string{"listUsers", "getUser"})
}

func TestExcludePathGlob(t *testing.T) {
	got := ids(Policy{Excludes: []Matcher{PathGlob("/admin/*")}}.Apply(ops()))
	assertEqual(t, got, []string{"listUsers", "getUser", "createPost"})
}

func TestCombinedReadOnlyPlusInclude(t *testing.T) {
	// read-only AND tagged "public": deleteUser is dropped twice over.
	p := Policy{ReadOnly: true, Includes: []Matcher{Tag("users")}}
	assertEqual(t, ids(p.Apply(ops())), []string{"listUsers", "getUser"})
}

func TestExcludeWinsOverInclude(t *testing.T) {
	p := Policy{Includes: []Matcher{Tag("users")}, Excludes: []Matcher{OperationID("getUser")}}
	assertEqual(t, ids(p.Apply(ops())), []string{"listUsers"})
}

func assertEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
