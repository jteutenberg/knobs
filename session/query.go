package session

import (
	"fmt"

	"github.com/jteutenberg/understate/core"
	"github.com/jteutenberg/understate/pathing"
)

func (s *Session) handleQuery(query []*core.Predicate, frame *core.Frame) error {
	fmt.Fprintln(s.Out, "Query: ")
	for _, p := range query {
		fmt.Fprintln(s.Out, " - ", p.String())
	}
	if len(query) == 0 {
		return nil
	}
	if len(query) == 1 {
		return s.handleSingleQuery(query[0], frame)
	}
	return s.handleConjunction(query, frame)
}

func (s *Session) handleSingleQuery(p *core.Predicate, frame *core.Frame) error {
	if act := s.Actions.GetAction(p); act != nil {
		return s.handleActionQuery(act)
	}
	if p.Definition == pathing.ShortestPathPredicate {
		return s.handleShortestPath(p)
	}
	return s.handlePredicateQuery(p, frame)
}

func (s *Session) handlePredicateQuery(p *core.Predicate, frame *core.Frame) error {
	answers := s.KB.Answer(p, frame, core.NewQueryContext())
	for ans := range answers {
		fmt.Fprintln(s.Out, "  -> ", ans.String())
	}
	fmt.Fprintln(s.Out, "Done.")
	return nil
}

func (s *Session) handleConjunction(query []*core.Predicate, frame *core.Frame) error {
	answers := core.AnswerConjunction(s.KB, query, frame, core.NewQueryContext())
	for ans := range answers {
		fmt.Fprintln(s.Out, "  ->")
		for _, p := range ans {
			fmt.Fprintln(s.Out, "    ", p.String())
		}
	}
	return nil
}
