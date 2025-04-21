package req

import "fmt"

var (
	ErrRoomDoesNotExist = fmt.Errorf("room does not exist")
	ErrRoomExists       = fmt.Errorf("room exists")
)
