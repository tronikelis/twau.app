package req

import "fmt"

var (
	ErrRoomDoesNotExist = fmt.Errorf("room does not exist")
	ErrUnknownAction    = fmt.Errorf("unknown action")
	ErrRoomExists       = fmt.Errorf("room exists")
	ErrNotYourTurn      = fmt.Errorf("not your turn")
)
