package rooms

import (
	"fmt"

	"twau.app/pkgs/server/req"
)

func postIndex(ctx req.ReqContext) error {
	playerName := ctx.Req().PostFormValue("player_name")

	_, roomId := ctx.Rooms.CreateRoom()

	_, err := ctx.Player()
	if err != nil {
		if err := ctx.SetPlayer(playerName); err != nil {
			return err
		}
	}

	ctx.Writer().Header().Set("hx-redirect", fmt.Sprintf("/rooms/%s", roomId))

	return nil
}
