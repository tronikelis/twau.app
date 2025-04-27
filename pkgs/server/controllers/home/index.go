package home

import (
	"twau.app/pkgs/server/req"
)

func getIndex(ctx req.ReqContext) error {
	player, _ := ctx.Player()

	return pageIndex(player.Name).Render(ctx.Context(), ctx.Writer())
}
