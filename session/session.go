package session

import (
	"bufio"
	"fmt"
	"io"

	"github.com/jteutenberg/bitset-go"
	"github.com/jteutenberg/understate/actions"
	"github.com/jteutenberg/understate/calculator"
	"github.com/jteutenberg/understate/core"
	ustio "github.com/jteutenberg/understate/io"
	"github.com/jteutenberg/understate/knowledgebase"
	"github.com/jteutenberg/understate/pathing"
	"github.com/jteutenberg/understate/rules"
)

// Session holds a knowledge base and the extra machinery needed to handle
// actions and shortest-path queries, matching the setup in understate's
// parser_test doParseExamples.
type Session struct {
	KB      *knowledgebase.KnowledgeBase
	Actions *actions.ActionSet
	Search  *pathing.Search
	Out     io.Writer
	Err     io.Writer
}

func New(out, errOut io.Writer) *Session {
	kb := knowledgebase.NewKnowledgeBase()
	ruleMachine := rules.NewRuleMachine(kb, kb.State)
	kb.AddAnswerer(ruleMachine)
	kb.AddAnswerer(calculator.NewCalculator(kb.State))
	kb.AddPredicateDefinition(calculator.Gt)
	kb.AddPredicateDefinition(calculator.Sum)
	kb.AddPredicateDefinition(pathing.ShortestPathPredicate)
	addRelationPredicates(kb)

	return &Session{
		KB:      kb,
		Actions: actions.NewActionSet(),
		Search:  pathing.NewSearch(kb),
		Out:     out,
		Err:     errOut,
	}
}

func addRelationPredicates(kb *knowledgebase.KnowledgeBase) {
	person := &core.Type{
		Name:    "Person",
		Atomics: bitset.NewIntSet(),
	}
	kb.AddPredicateDefinition(&core.PredicateDefinition{
		Functor: "parent",
		ArgDefinitions: []core.ArgumentDefinition{
			{Label: "Parent", Type: person},
			{Label: "Child", Type: person},
		},
	})
	kb.AddPredicateDefinition(&core.PredicateDefinition{
		Functor: "sibling",
		ArgDefinitions: []core.ArgumentDefinition{
			{Label: "A", Type: person},
			{Label: "B", Type: person},
		},
	})
	kb.AddPredicateDefinition(&core.PredicateDefinition{
		Functor: "grandparent",
		ArgDefinitions: []core.ArgumentDefinition{
			{Label: "Grandparent", Type: person},
			{Label: "Grandchild", Type: person},
		},
	})
}

func (s *Session) Run(r io.Reader) error {
	parser := ustio.NewPredicateReader(
		[]byte{knowledgebase.ActionSeparator, knowledgebase.RuleSeparator},
		[]byte{knowledgebase.AssertTerminator, knowledgebase.CommandTerminator, knowledgebase.QueryTerminator},
	)
	for token := range parser.Parse(bufio.NewReader(r)) {
		if err := s.Handle(token); err != nil {
			fmt.Fprintln(s.Err, err)
		}
	}
	return nil
}

// Handle routes one parsed token to a type-specific handler.
// Assertions, predicate definitions, and rules are applied inside Process
// and produce no further work here.
func (s *Session) Handle(token ustio.ParseResult) error {
	query, command, actionDef, frame, err := s.KB.Process(token)
	if err != nil {
		return err
	}
	switch {
	case query != nil:
		return s.handleQuery(query, frame)
	case command != nil:
		return s.handleCommand(command)
	case actionDef != nil:
		return s.handleActionDefinition(actionDef)
	default:
		return nil
	}
}
