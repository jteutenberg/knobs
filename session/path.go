package session

import (
	"fmt"
	"strings"

	"github.com/jteutenberg/understate/core"
)

func (s *Session) handleShortestPath(p *core.Predicate) (<-chan string, error) {
	from, okFrom := p.GetArgument(0).(*core.Atomic)
	to, okTo := p.GetArgument(1).(*core.Atomic)
	connectorPred, okConn := p.GetArgument(2).(*core.Predicate)
	if !okFrom || !okTo || !okConn {
		return nil, fmt.Errorf("shortestPath expects atomic from/to and a connector predicate")
	}
	path := s.Search.ShortestPath(from, to, connectorPred.Definition)
	if path == nil {
		fmt.Fprintln(s.Err, "No path between", from.Value, "and", to.Value)
		return singleAnswer("None"), nil
	}
	parts := make([]string, len(path))
	for i, v := range path {
		parts[i] = v.Value
	}
	return singleAnswer(strings.Join(parts, ", ")), nil
}
