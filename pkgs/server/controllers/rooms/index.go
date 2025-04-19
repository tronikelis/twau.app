package rooms

import (
	"fmt"
	"net/http"

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

	playerCookies, err := req.GetPlayerCookies(ctx.Req(), ctx.SecretKey)
	if err != nil {
		playerCookies, err = req.NewPlayerCookies(playerName, ctx.SecretKey)
		if err != nil {
			return err
		}

		http.SetCookie(ctx.Writer(), playerCookies.Id)
		http.SetCookie(ctx.Writer(), playerCookies.Name)
	}

	ctx.Writer().Header().Set("hx-redirect", fmt.Sprintf("/rooms/%s", roomId))

	return nil
}
