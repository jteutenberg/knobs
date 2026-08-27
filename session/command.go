package session

import (
	"fmt"

	"github.com/jteutenberg/understate/core"
)

func (s *Session) handleCommand(command *core.Predicate) (<-chan string, error) {
	fmt.Fprintln(s.Err, "Command: ", command.String())
	act := s.Actions.GetAction(command)
	if act == nil {
		fmt.Fprintln(s.Err, "Action not found")
		return nil, nil
	}
	fmt.Fprintln(s.Err, act.Signature.String())
	if !act.IsApplicable(s.KB) {
		fmt.Fprintln(s.Err, "Action is not applicable")
	}
	act.ApplyTo(s.KB.State)
	return nil, nil
}
