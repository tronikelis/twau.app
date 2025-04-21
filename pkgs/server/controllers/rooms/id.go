package rooms

import (
	"fmt"
	"log"
	"net/http"

	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/server/req"

	"github.com/gorilla/websocket"
)

func postId(ctx req.ReqContext) error {
	roomId := ctx.Req().PathValue("id")

	playerName := ctx.Req().PostFormValue("player_name")

	_, ok := ctx.Rooms.Room(roomId)
	if !ok {
		return req.ErrRoomDoesNotExist
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

var wsUpgrader = websocket.Upgrader{}

func withLog(fn func() error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("withLog recovered", err)
		}
	}()

	if err := fn(); err != nil {
		log.Println("withLog", err)
	}
}

func handleWsId(
	ctx req.ReqContext,
	conn *websocket.Conn,
	playerCookies req.PlayerCookies,
	room *game_state.Room,
	roomId string,
) error {
	log.Println(fmt.Sprintf("%s connected, [%s]", playerCookies.Name.Value, playerCookies.Id.Value))

	room.AddPlayer(conn, game_state.NewPlayer(playerCookies.Id.Value, playerCookies.Name.Value))
	defer room.State(func(state game_state.GameState) {
		if state.GetGame().PlayersOnline() == 0 {
			ctx.Rooms.QueueDelete(roomId)
		}
	})
	defer room.RemovePlayer(conn, playerCookies.Id.Value)
	defer conn.Close()

	return room.GameLoop(conn, playerCookies.Id.Value)
}

func wsId(ctx req.ReqContext) error {
	roomId := ctx.Req().PathValue("id")

	if !ctx.Rooms.HasRoom(roomId) {
		return req.ErrRoomDoesNotExist
	}

	playerCookies, err := req.GetPlayerCookies(ctx.Req(), ctx.SecretKey)
	if err != nil {
		return err
	}

	room, ok := ctx.Rooms.Room(roomId)
	if !ok {
		return req.ErrRoomDoesNotExist
	}

	socket, err := wsUpgrader.Upgrade(ctx.Writer(), ctx.Req(), nil)
	if err != nil {
		return err
	}

	go withLog(func() error {
		return handleWsId(ctx, socket, playerCookies, room, roomId)
	})

	return nil
}
