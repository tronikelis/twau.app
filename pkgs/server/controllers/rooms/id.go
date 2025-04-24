package rooms

import (
	"fmt"
	"log"
	"net/http"

	"twau.app/pkgs/game_state"
	"twau.app/pkgs/server/req"
	"twau.app/pkgs/ws"

	"github.com/gorilla/websocket"
)

func postId(ctx req.ReqContext) error {
	roomId := ctx.Req().PathValue("id")

	playerName := ctx.Req().PostFormValue("player_name")

	_, ok := ctx.Rooms.Room(roomId)
	if !ok {
		return req.ErrRoomDoesNotExist
	}

	_, err := ctx.Player()
	if err != nil {
		if err := ctx.SetPlayer(playerName); err != nil {
			return err
		}
	}

	ctx.Writer().Header().Set("hx-redirect", fmt.Sprintf("/rooms/%s", roomId))

	return nil
}

func getId(ctx req.ReqContext) error {
	roomId := ctx.Req().PathValue("id")

	if !ctx.Rooms.HasRoom(roomId) {
		return req.ErrRoomDoesNotExist
	}

	_, err := ctx.Req().Cookie(req.CookiePlayerName.Name)
	if err != nil {
		return pagePlayerCreate(roomId).Render(ctx.Context(), ctx.Writer())
	}
	_, err = ctx.Req().Cookie(req.CookiePlayerId.Name)
	if err != nil {
		return pagePlayerCreate(roomId).Render(ctx.Context(), ctx.Writer())
	}

	return pageRoomId(roomId).Render(ctx.Context(), ctx.Writer())
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWsId(
	ctx req.ReqContext,
	conn *ws.ConnSafe,
	player req.Player,
	room *game_state.Room,
	roomId string,
) error {
	log.Println(fmt.Sprintf("%s connected, [%s]", player.Name, player.Id))

	room.AddPlayer(conn, game_state.NewPlayer(player.Id, player.Name))
	defer room.State(func(state game_state.GameState) {
		if state.GetGame().PlayersOnline() == 0 {
			ctx.Rooms.QueueDelete(roomId)
		}
	})
	defer room.RemovePlayer(conn, player.Id)
	defer conn.Close()

	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		if err := room.GameLoop(bytes, player.Id); err != nil {
			return err
		}
	}
}

func wsId(ctx req.ReqContext) error {
	roomId := ctx.Req().PathValue("id")

	if !ctx.Rooms.HasRoom(roomId) {
		return req.ErrRoomDoesNotExist
	}

	player, err := ctx.Player()
	if err != nil {
		return err
	}

	room, ok := ctx.Rooms.Room(roomId)
	if !ok {
		return req.ErrRoomDoesNotExist
	}

	conn, err := wsUpgrader.Upgrade(ctx.Writer(), ctx.Req(), nil)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("wsId recover", "err", err)
			}
		}()

		if err := handleWsId(ctx, ws.NewConnSafe(conn), player, room, roomId); err != nil {
			log.Println("wsId", "err", err)
		}
	}()

	return nil
}
