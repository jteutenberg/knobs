package session

import (
	"fmt"

	"github.com/jteutenberg/understate/actions"
)

func (s *Session) handleActionDefinition(actionDef *actions.Action) error {
	fmt.Fprintln(s.Out, "Action definition: ", actionDef.Signature.String())
	s.Actions.AddAction(actionDef)
	return nil
}

func (s *Session) handleActionQuery(act *actions.Action) error {
	fmt.Fprintln(s.Out, "Action query: ", act.Signature.String())
	fmt.Fprintln(s.Out, " Applicable: ", act.IsApplicable(s.KB))
	return nil
}
