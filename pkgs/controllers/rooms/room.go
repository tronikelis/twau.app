package rooms

import (
	"context"
	"fmt"
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
		panic("room does not exist")
	}

	ws, err := wsUpgrader.Upgrade(ctx.Writer(), ctx.Req(), nil)
	if err != nil {
		panic(err)
	}

	room.State(func(state game_state.GameState) error {
		state.Game().AddPlayer(game_state.NewPlayer(playerId.Value, playerName.Value))
		return nil
	})
	defer room.State(func(state game_state.GameState) error { // 3. sync changes to others
		state.Game().RemovePlayer(playerId.Value)
		players := state.Game().Players()

		if err := room.WsRoom.WriteAll(func(writer io.Writer) error {
			return partialPlayers(players).Render(context.Background(), writer)
		}); err != nil {
			fmt.Println(err)
		}

		return nil
	})

	room.WsRoom.Add(ws)
	defer room.WsRoom.Delete(ws) // 2. remove from ws room
	defer ws.Close()             // 1. close the ws conn

	err = room.State(func(state game_state.GameState) error {
		players := state.Game().Players()

		if err := room.WsRoom.WriteAll(func(writer io.Writer) error {
			return partialPlayers(players).Render(context.Background(), writer)
		}); err != nil {
			fmt.Println(err)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
