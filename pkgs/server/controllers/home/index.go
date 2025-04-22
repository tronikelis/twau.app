package home

import (
	"twau.app/pkgs/server/req"
)

func getIndex(ctx req.ReqContext) error {
	var playerName string
	playerNameCookie, err := ctx.Req().Cookie(req.CookiePlayerName.Name)
	if err == nil {
		playerName = playerNameCookie.Value
	}

	return pageIndex(playerName).Render(ctx.Context(), ctx.Writer())
}
