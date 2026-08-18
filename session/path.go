package session

import (
	"fmt"

	"github.com/jteutenberg/understate/core"
)

func (s *Session) handleShortestPath(p *core.Predicate) error {
	from, okFrom := p.GetArgument(0).(*core.Atomic)
	to, okTo := p.GetArgument(1).(*core.Atomic)
	connectorPred, okConn := p.GetArgument(2).(*core.Predicate)
	if !okFrom || !okTo || !okConn {
		return fmt.Errorf("shortestPath expects atomic from/to and a connector predicate")
	}
	path := s.Search.ShortestPath(from, to, connectorPred.Definition)
	if path == nil {
		fmt.Fprintln(s.Out, "No path between", from.Value, "and", to.Value)
		return nil
	}
	for i, v := range path {
		if i > 0 {
			fmt.Fprint(s.Out, " -> ")
		}
		fmt.Fprint(s.Out, v.Value)
	}
	fmt.Fprintln(s.Out)
	return nil
}
