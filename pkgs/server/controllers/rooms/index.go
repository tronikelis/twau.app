package rooms

import (
	"fmt"

	"twau.app/pkgs/server/req"
)

func postIndex(ctx req.ReqContext) error {
	playerName := ctx.Req().PostFormValue("player_name")
	roomPassword := ctx.Req().PostFormValue("room_password")

	_, roomId := ctx.Rooms.CreateRoom(roomPassword)

	_, err := ctx.Player()
	if err != nil {
		if err := ctx.SetPlayer(playerName); err != nil {
			return err
		}
	}

	ctx.Writer().Header().Set("hx-redirect", fmt.Sprintf("/rooms/%s", roomId))

	roomPasswordCookie := req.CookieRoomPassword
	roomPasswordCookie.Value = roomPassword
	ctx.SetCookie(&roomPasswordCookie)

	return nil
}
