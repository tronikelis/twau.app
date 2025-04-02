package rooms

import (
	"context"
	"io"
	"net/http"

	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/server/req"

	"github.com/gorilla/websocket"
	"github.com/tronikelis/maruchi"
)

var wsUpgrader = websocket.Upgrader{}

func getRoomId(ctx maruchi.ReqContext) {
	reqContext := req.GetReqContext(ctx)

	roomId := ctx.Req().PathValue("id")

	if !reqContext.Rooms.HasRoom(roomId) {
		ctx.Writer().WriteHeader(http.StatusNotFound)
		return
	}

	// playerName, err := ctx.Req().Cookie(req.CookiePlayerName.Name)
	// if err != nil {
	// 	return pagePlayerCreate()
	// }
	// playerId, err := ctx.Req().Cookie(req.CookiePlayerId.Name)
	// if err != nil {
	// 	return pagePlayerCreate()
	// }

	pageRoomId(roomId).Render(ctx.Context(), ctx.Writer())
}

func wsRoomId(ctx maruchi.ReqContext) {
	reqContext := req.GetReqContext(ctx)

	roomId := ctx.Req().PathValue("id")

	if !reqContext.Rooms.HasRoom(roomId) {
		ctx.Writer().WriteHeader(http.StatusNotFound)
		return
	}

	playerId, err := ctx.Req().Cookie(req.CookiePlayerId.Name)
	if err != nil {
		panic(err)
	}

	playerName, err := ctx.Req().Cookie(req.CookiePlayerName.Name)
	if err != nil {
		panic(err)
	}

	room, ok := reqContext.Rooms.Room(roomId)
	if !ok {
		panic("room already exists")
	}

	ws, err := wsUpgrader.Upgrade(ctx.Writer(), ctx.Req(), nil)
	if err != nil {
		panic(err)
	}
	defer ws.Close()

	room.WsRoom.Add(ws)
	defer room.WsRoom.Delete(ws)

	room.State.Game().AddPlayer(game_state.NewPlayer(playerId.Value, playerName.Value))
	defer room.State.Game().RemovePlayer(playerId.Value)

	// sync up new connections
	switch state := room.State.(type) {
	case *game_state.Game:
		players := state.Players()

		if err := room.WsRoom.WriteAll(func(writer io.Writer) error {
			return partialPlayers(players).Render(context.Background(), writer)
		}); err != nil {
			panic(err)
		}
	default:
		panic("unsupported game state")
	}

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			panic(err)
		}
	}
}
