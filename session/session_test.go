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
	return out.String()
}

func TestPredicateQuery(t *testing.T) {
	out := runInput(t, `
:parent(Parent:Person, Child:Person).
parent(sam, alex).
parent(jo, alex).
parent(X, alex)?
n
n
`)
	if !strings.Contains(out, "parent(sam, alex)") || !strings.Contains(out, "parent(jo, alex)") {
		t.Fatalf("expected both parent answers, got:\n%s", out)
	}
	if !strings.Contains(out, "Done") {
		t.Fatalf("expected Done after last next, got:\n%s", out)
	}
}

func TestPredicateQueryWaitsForNext(t *testing.T) {
	out := runInput(t, `
:parent(Parent:Person, Child:Person).
parent(sam, alex).
parent(jo, alex).
parent(X, alex)?
`)
	if strings.Count(out, "parent(") != 1 {
		t.Fatalf("expected only the first answer without next, got:\n%s", out)
	}
	if strings.Contains(out, "Done") {
		t.Fatalf("did not expect Done without a next request, got:\n%s", out)
	}
}

func TestNextIsCaseInsensitive(t *testing.T) {
	out := runInput(t, `
:parent(Parent:Person, Child:Person).
parent(sam, alex).
parent(jo, alex).
parent(X, alex)?
N
n
`)
	if !strings.Contains(out, "parent(sam, alex)") || !strings.Contains(out, "parent(jo, alex)") {
		t.Fatalf("expected both parent answers with N/n, got:\n%s", out)
	}
	if !strings.Contains(out, "Done") {
		t.Fatalf("expected Done after last next, got:\n%s", out)
	}
}

func TestConjunctionQuery(t *testing.T) {
	out := runInput(t, `
:parent(Parent:Person, Child:Person).
parent(sam, alex).
parent(sam, sal).
parent(X, alex),
parent(X, sal)?
n
`)
	if !strings.Contains(out, "parent(sam, alex)") || !strings.Contains(out, "parent(sam, sal)") {
		t.Fatalf("expected both conjuncts, got:\n%s", out)
	}
	if !strings.Contains(out, "Done") {
		t.Fatalf("expected Done after last next, got:\n%s", out)
	}
}

func TestCommandDoesNotWaitForNext(t *testing.T) {
	out := runInput(t, `
:adjacent(A:location, B:location).
:link(X:location, Y:location).
adjacent(a, b).
link(X, Y)
|
  not(adjacent(X, Y)),
  not(eq(X, Y))
| adjacent(X, Y).
link(b, c)!
adjacent(b, c)?
n
`)
	if !strings.Contains(out, "adjacent(b, c)") {
		t.Fatalf("expected adjacent fact after command without an intervening next, got:\n%s", out)
	}
	if !strings.Contains(out, "Done") {
		t.Fatalf("expected Done after query next, got:\n%s", out)
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
n
link(b, c)!
adjacent(b, c)?
n
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected action query to answer true, got:\n%s", out)
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
n
`)
	if !strings.Contains(out, "a, b, c") && !strings.Contains(out, "c, b, a") {
		t.Fatalf("expected a path between a and c, got:\n%s", out)
	}
	if !strings.Contains(out, "Done") {
		t.Fatalf("expected Done after last next, got:\n%s", out)
	}
}
