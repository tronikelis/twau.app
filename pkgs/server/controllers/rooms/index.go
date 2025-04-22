package rooms

import (
	"fmt"

	"word-amongus-game/pkgs/random"
	"word-amongus-game/pkgs/server/req"
)

func postIndex(ctx req.ReqContext) error {
	playerName := ctx.Req().PostFormValue("player_name")

	roomId, err := random.RandomHex(random.LengthRoomId)
	if err != nil {
		return err
	}

	_, ok := ctx.Rooms.CreateRoom(roomId)
	if !ok {
		return req.ErrRoomExists
	}

	_, err = ctx.Player()
	if err != nil {
		if err := ctx.SetPlayer(playerName); err != nil {
			return err
		}
	}

	ctx.Writer().Header().Set("hx-redirect", fmt.Sprintf("/rooms/%s", roomId))

	return nil
}
