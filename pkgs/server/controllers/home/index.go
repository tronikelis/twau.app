package home

import (
	"twau.app/pkgs/server/req"
)

func getIndex(ctx req.ReqContext) error {
	player, _ := ctx.Player()

	return pageIndex(player.Name).Render(ctx.Context(), ctx.Writer())
}

func allHxEditPlayerName(ctx req.ReqContext) error {
	player, _ := ctx.Player()

	switch ctx.Req().Method {
	case "GET":
		return partialEditPlayerName(player.Name).Render(ctx.Context(), ctx.Writer())
	case "PUT":
		if err := ctx.Req().ParseForm(); err != nil {
			return err
		}
		playerName := ctx.Req().PostForm.Get("player_name")

		if err := ctx.SetPlayer(playerName); err != nil {
			return err
		}

		return partialPlayerName(playerName).Render(ctx.Context(), ctx.Writer())
	default:
		return req.ErrMethodNotAllowed
	}
}
