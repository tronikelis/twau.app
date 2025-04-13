package home

import (
	"word-amongus-game/pkgs/server/req"
)

func getIndex(ctx req.ReqContext) error {
	return pageIndex().Render(ctx.Context(), ctx.Writer())
}
