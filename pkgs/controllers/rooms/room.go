package rooms

import (
	"net/http"

	"word-amongus-game/pkgs/server/req"

	"github.com/gorilla/websocket"
	"github.com/tronikelis/maruchi"
)

var wsUpgrader = websocket.Upgrader{}

func getRoomId(ctx maruchi.ReqContext) {
	gameId := ctx.Req().PathValue("id")

	if !req.GetReqContext(ctx).States.HasGame(gameId) {
		ctx.Writer().WriteHeader(http.StatusNotFound)
		return
	}

	ctx.Writer().Write([]byte(gameId))
}

func wsRoomId(ctx maruchi.ReqContext) {
	reqContext := req.GetReqContext(ctx)

	playerId, err := ctx.Req().Cookie(req.CookiePlayerId.Name)
	if err != nil {
		panic(err)
	}

	if _, ok := reqContext.WsByPlayerId.Load(playerId.Value); ok {
		panic("playerId taken")
	}

	ws, err := wsUpgrader.Upgrade(ctx.Writer(), ctx.Req(), nil)
	if err != nil {
		panic(err)
	}
	defer ws.Close()

	reqContext.WsByPlayerId.Insert(playerId.Value, ws)
	defer reqContext.WsByPlayerId.Delete(playerId.Value)

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			panic(err)
		}
	}
}
