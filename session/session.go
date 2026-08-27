package session

import (
	"bufio"
	"fmt"
	"io"
	"strings"

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

	return &Session{
		KB:      kb,
		Actions: actions.NewActionSet(),
		Search:  pathing.NewSearch(kb),
		Out:     out,
		Err:     errOut,
	}
}

type inputEvent struct {
	next  bool
	token ustio.ParseResult
}

func (s *Session) Run(r io.Reader) error {
	parser := ustio.NewPredicateReader(
		[]byte{knowledgebase.ActionSeparator, knowledgebase.RuleSeparator},
		[]byte{knowledgebase.AssertTerminator, knowledgebase.CommandTerminator, knowledgebase.QueryTerminator},
	)
	var answering <-chan string
	writeNext := func() {
		if answering == nil {
			return
		}
		ans, ok := <-answering
		if !ok {
			fmt.Fprintln(s.Out, "Done")
			answering = nil
			return
		}
		fmt.Fprintln(s.Out, ans)
	}
	abandon := func() {
		if answering == nil {
			return
		}
		ch := answering
		answering = nil
		go func() {
			for range ch {
			}
		}()
	}

	for event := range readInputs(parser, bufio.NewReader(r)) {
		if event.next {
			writeNext()
			continue
		}
		abandon()
		ch, err := s.Handle(event.token)
		if err != nil {
			fmt.Fprintln(s.Err, err)
			continue
		}
		if ch == nil {
			continue
		}
		answering = ch
		writeNext()
	}
	return nil
}

// Handle routes one parsed token to a type-specific handler.
// Assertions, predicate definitions, and rules are applied inside Process
// and produce no further work here.
func (s *Session) Handle(token ustio.ParseResult) (<-chan string, error) {
	query, command, actionDef, frame, err := s.KB.Process(token)
	if err != nil {
		return nil, err
	}
	switch {
	case query != nil:
		return s.handleQuery(query, frame)
	case command != nil:
		return s.handleCommand(command)
	case actionDef != nil:
		return s.handleActionDefinition(actionDef)
	default:
		return nil, nil
	}
}

func singleAnswer(s string) <-chan string {
	ch := make(chan string, 1)
	ch <- s
	close(ch)
	return ch
}

func predicateString(query *core.Predicate, answer *core.Predicate, frame *core.Frame, makeVars bool) string {
	if !makeVars {
		return answer.String()
	}
	fr := frame.Clone()
	p := query.CloneInFrame(fr)
	p.Unify(answer)
	varStrings := make(map[string]string)
	for i, varRef := range p.VarRefs {
		if _, ok := frame.Vars[varRef.Label]; ok && frame.Vars[varRef.Label].Ref == nil {
			arg := p.GetArgument(i)
			varStrings[varRef.Label] = arg.(*core.Atomic).String()
		}
	}
	// Return the contents of varStrings as "key1=value1, key2=value2, ..."
	pairs := make([]string, 0, len(varStrings))
	for k, v := range varStrings {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ", ")
}
func predicatesString(query []*core.Predicate, answer []*core.Predicate, frame *core.Frame, makeVars bool) string {
	if !makeVars {
		answers := make([]string, 0, len(answer))
		for _, p := range answer {
			answers = append(answers, p.String())
		}
		return strings.Join(answers, ", ")
	}
	varStrings := make(map[string]string)
	fr := frame.Clone()
	for j, q := range query {
		p := q.CloneInFrame(fr)
		p.Unify(answer[j])
		for i, varRef := range p.VarRefs {
			if _, ok := frame.Vars[varRef.Label]; ok && frame.Vars[varRef.Label].Ref == nil {
				arg := p.GetArgument(i)
				varStrings[varRef.Label] = arg.(*core.Atomic).String()
			}
		}
	}
	// Return the contents of varStrings as "key1=value1, key2=value2, ..."
	pairs := make([]string, 0, len(varStrings))
	for k, v := range varStrings {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ", ")
}

func readInputs(pr *ustio.PredicateReader, reader io.ByteReader) <-chan inputEvent {
	out := make(chan inputEvent)
	go func() {
		defer close(out)
		line := make([]byte, 0, 10000)
		inComment := false
		token := ustio.ParseResult{
			Predicates: make([]string, 0, 5),
			Separators: make([]byte, 0, 5),
		}
		for {
			b, err := reader.ReadByte()
			if err != nil {
				if isNextToken(line) {
					out <- inputEvent{next: true}
				}
				return
			}
			inComment = inComment || (b == '#')
			if pr.Whitespace[b] || inComment {
				if b == '\r' || b == '\n' {
					inComment = false
				}
				if pr.Whitespace[b] && isNextToken(line) {
					out <- inputEvent{next: true}
					line = line[:0]
				}
				continue
			}
			line = append(line, b)
			if !pr.Terminators[b] && !pr.Seperators[b] {
				continue
			}
			sep := line[len(line)-1]
			token.Predicates = append(token.Predicates, string(line[:len(line)-1]))
			line = line[:0]
			if pr.Terminators[sep] {
				token.Terminator = sep
				out <- inputEvent{token: token}
				token = ustio.ParseResult{
					Predicates: make([]string, 0, 5),
					Separators: make([]byte, 0, 5),
				}
			} else {
				token.Separators = append(token.Separators, sep)
			}
		}
	}()
	return out
}

func isNextToken(line []byte) bool {
	return string(line) == "n" || string(line) == "N"
}
