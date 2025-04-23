package game_state

import "fmt"

var (
	ErrNotYourTurn   = fmt.Errorf("not your turn")
	ErrBadAction     = fmt.Errorf("bad action")
	ErrUnknownAction = fmt.Errorf("unknown action")
)
