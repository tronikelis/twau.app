package rooms

import (
	"net/http"

	"word-amongus-game/pkgs/server/req"

	"github.com/tronikelis/maruchi"
)

func getRoomId(ctx maruchi.ReqContext) {
	gameId := ctx.Req().PathValue("id")

	if !req.GetStates(ctx).HasGame(gameId) {
		ctx.Writer().WriteHeader(http.StatusNotFound)
		return
	}

	ctx.Writer().Write([]byte(gameId))
}
