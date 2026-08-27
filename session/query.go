package session

import (
	"fmt"

	"github.com/jteutenberg/understate/core"
	"github.com/jteutenberg/understate/pathing"
)

func (s *Session) handleQuery(query []*core.Predicate, frame *core.Frame) (<-chan string, error) {
	fmt.Fprintln(s.Err, "Query: ")
	for _, p := range query {
		fmt.Fprintln(s.Err, " - ", p.String())
	}
	if len(query) == 0 {
		return nil, nil
	}
	if len(query) == 1 {
		return s.handleSingleQuery(query[0], frame)
	}
	return s.handleConjunction(query, frame)
}

func (s *Session) handleSingleQuery(p *core.Predicate, frame *core.Frame) (<-chan string, error) {
	if act := s.Actions.GetAction(p); act != nil {
		return s.handleActionQuery(act)
	}
	if p.Definition == pathing.ShortestPathPredicate {
		return s.handleShortestPath(p)
	}
	return s.handlePredicateQuery(p, frame)
}

func (s *Session) handlePredicateQuery(p *core.Predicate, frame *core.Frame) (<-chan string, error) {
	answers := s.KB.Answer(p, frame, core.NewQueryContext())
	out := make(chan string)
	go func() {
		defer close(out)
		for ans := range answers {
			out <- predicateString(p, ans, frame, true)
		}
	}()
	return out, nil
}

func (s *Session) handleConjunction(query []*core.Predicate, frame *core.Frame) (<-chan string, error) {
	answers := core.AnswerConjunction(s.KB, query, frame, core.NewQueryContext())
	out := make(chan string)
	go func() {
		defer close(out)
		for ans := range answers {
			out <- predicatesString(query, ans, frame, true)
		}
	}()
	return out, nil
}
