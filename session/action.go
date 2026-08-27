package session

import (
	"fmt"
	"strconv"

	"github.com/jteutenberg/understate/actions"
)

func (s *Session) handleActionDefinition(actionDef *actions.Action) (<-chan string, error) {
	fmt.Fprintln(s.Err, "Action definition: ", actionDef.Signature.String())
	s.Actions.AddAction(actionDef)
	return nil, nil
}

func (s *Session) handleActionQuery(act *actions.Action) (<-chan string, error) {
	fmt.Fprintln(s.Err, "Action query: ", act.Signature.String())
	return singleAnswer(strconv.FormatBool(act.IsApplicable(s.KB))), nil
}
