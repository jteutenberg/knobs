package session

import (
	"bytes"
	"strings"
	"testing"
)

func runInput(t *testing.T, input string) string {
	t.Helper()
	var out, errOut bytes.Buffer
	s := New(&out, &errOut)
	if err := s.Run(strings.NewReader(input)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if errOut.Len() > 0 {
		t.Fatalf("unexpected errors: %s", errOut.String())
	}
	return out.String()
}

func TestPredicateQuery(t *testing.T) {
	out := runInput(t, `
parent(sam, alex).
parent(jo, alex).
parent(X, alex)?
`)
	if !strings.Contains(out, "Query:") {
		t.Fatalf("expected query header, got:\n%s", out)
	}
	if !strings.Contains(out, "parent(sam, alex)") || !strings.Contains(out, "parent(jo, alex)") {
		t.Fatalf("expected both parent answers, got:\n%s", out)
	}
}

func TestConjunctionQuery(t *testing.T) {
	out := runInput(t, `
parent(sam, alex).
parent(sam, sal).
parent(X, alex),
parent(X, sal)?
`)
	if !strings.Contains(out, "  ->") {
		t.Fatalf("expected conjunction answers, got:\n%s", out)
	}
	if !strings.Contains(out, "parent(sam, alex)") || !strings.Contains(out, "parent(sam, sal)") {
		t.Fatalf("expected both conjuncts, got:\n%s", out)
	}
}

func TestCommandAndAction(t *testing.T) {
	out := runInput(t, `
:adjacent(A:location, B:location).
:link(X:location, Y:location).
adjacent(a, b).
link(X, Y)
|
  not(adjacent(X, Y)),
  not(eq(X, Y))
| adjacent(X, Y).
link(b, c)?
link(b, c)!
adjacent(b, c)?
`)
	if !strings.Contains(out, "Action definition:") {
		t.Fatalf("expected action definition, got:\n%s", out)
	}
	if !strings.Contains(out, "Action query:") {
		t.Fatalf("expected action query, got:\n%s", out)
	}
	if !strings.Contains(out, "Command:") || !strings.Contains(out, "Action is applicable") {
		t.Fatalf("expected applicable command, got:\n%s", out)
	}
	if !strings.Contains(out, "adjacent(b, c)") {
		t.Fatalf("expected adjacent fact after command, got:\n%s", out)
	}
}

func TestShortestPath(t *testing.T) {
	out := runInput(t, `
:adjacent(A:location, B:location).
adjacent(a, b).
adjacent(b, c).
shortestPath(a, c, adjacent(X, Y))?
`)
	if !strings.Contains(out, "c -> b -> a") && !strings.Contains(out, "a -> b -> c") {
		t.Fatalf("expected a path between a and c, got:\n%s", out)
	}
}
