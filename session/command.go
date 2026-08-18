package session

import (
	"fmt"

	"github.com/jteutenberg/understate/core"
)

func (s *Session) handleCommand(command *core.Predicate) error {
	fmt.Fprintln(s.Out, "Command: ", command.String())
	act := s.Actions.GetAction(command)
	if act == nil {
		fmt.Fprintln(s.Out, "Action not found")
		return nil
	}
	fmt.Fprintln(s.Out, "Action: ", act.Signature.String())
	if act.IsApplicable(s.KB) {
		fmt.Fprintln(s.Out, "Action is applicable")
	} else {
		fmt.Fprintln(s.Out, "Action is not applicable")
	}
	act.ApplyTo(s.KB.State)
	return nil
}
