package rooms

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"twau.app/pkgs/game_state"
	"twau.app/pkgs/server/req"
	"twau.app/pkgs/ws"

	"github.com/gorilla/websocket"
)

func postId(ctx req.ReqContext) error {
	roomId := ctx.Req().PathValue("id")

	playerName := ctx.Req().PostFormValue("player_name")
	roomPassword := ctx.Req().PostFormValue("room_password")

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

	roomPasswordCookie := req.CookieRoomPassword
	roomPasswordCookie.Value = roomPassword
	ctx.SetCookie(&roomPasswordCookie)

	ctx.Writer().Header().Set("hx-redirect", fmt.Sprintf("/rooms/%s", roomId))

	return nil
}

func getId(ctx req.ReqContext) error {
	roomId := ctx.Req().PathValue("id")

	room, ok := ctx.Rooms.Room(roomId)
	if !ok {
		return req.ErrRoomDoesNotExist
	}

	player, err := ctx.Player()
	if err != nil {
		return pagePlayerRoomJoin(roomId, "", room.Password() != "").Render(ctx.Context(), ctx.Writer())
	}

	roomPasswordCookie, _ := ctx.Cookie(req.CookieRoomPassword.Name)
	if roomPasswordCookie == nil {
		roomPasswordCookie = &http.Cookie{}
	}
	if room.Password() != "" && roomPasswordCookie.Value != room.Password() {
		return pagePlayerRoomJoin(roomId, player.Name, true).Render(ctx.Context(), ctx.Writer())
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
	defer room.State(func(state game_state.GameState) error {
		if state.GetGame().Players().Online() == 0 {
			ctx.Rooms.QueueDelete(roomId)
		}
		return nil
	})
	defer room.RemovePlayer(conn, player.Id)

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

	if room.Password() != "" {
		roomPasswordCookie, err := ctx.Cookie(req.CookieRoomPassword.Name)
		if err != nil {
			return err
		}
		if room.Password() != roomPasswordCookie.Value {
			return req.ErrRoomPasswordIncorrect
		}
	}

	conn, err := wsUpgrader.Upgrade(ctx.Writer(), ctx.Req(), nil)
	if err != nil {
		return err
	}
	connSafe := ws.NewConnSafe(conn)
	// maybe this logic should be in NewConnSafe?
	connCloseChan := make(chan struct{})

	go func() {
		defer connSafe.Close()

		for {
			select {
			case <-connCloseChan:
				return
			case <-time.After(time.Second * 10):
				if err := connSafe.WriteControl(websocket.PingMessage, nil, time.Now().Add(ws.WriteWait)); err != nil {
					log.Println("write ping", "err", err)
					return
				}
			}
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("wsId recover", "err", err)
			}
		}()
		defer connSafe.Close()
		defer close(connCloseChan)

		if err := handleWsId(ctx, connSafe, player, room, roomId); err != nil {
			log.Println("wsId", "err", err)
		}
	}()

	return nil
}
