package rooms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/server/req"
	"word-amongus-game/pkgs/ws"

	"github.com/gorilla/websocket"
	"github.com/tronikelis/maruchi"
)

func postId(ctx maruchi.ReqContext) {
	reqContext := req.GetReqContext(ctx)

	roomId := ctx.Req().PathValue("id")

	playerName := ctx.Req().PostFormValue("player_name")

	playerId, err := game_state.RandomHex()
	if err != nil {
		panic(err)
	}

	room, ok := reqContext.Rooms.Room(roomId)
	if !ok {
		panic("room does not exist")
	}

	if err := room.State(func(state game_state.GameState) error {
		state.Game().AddPlayer(game_state.NewPlayer(playerId, playerName))
		return nil
	}); err != nil {
		panic(err)
	}

	ctx.Writer().Header().Set("hx-redirect", fmt.Sprintf("/rooms/%s", roomId))

	playerIdCookie := req.CookiePlayerId
	playerIdCookie.Value = playerId

	playerNameCookie := req.CookiePlayerName
	playerNameCookie.Value = playerName

	http.SetCookie(ctx.Writer(), &playerIdCookie)
	http.SetCookie(ctx.Writer(), &playerNameCookie)
}

func getId(ctx maruchi.ReqContext) {
	reqContext := req.GetReqContext(ctx)

	roomId := ctx.Req().PathValue("id")

	if !reqContext.Rooms.HasRoom(roomId) {
		ctx.Writer().WriteHeader(http.StatusNotFound)
		return
	}

	_, err := ctx.Req().Cookie(req.CookiePlayerName.Name)
	if err != nil {
		pagePlayerCreate(roomId).Render(ctx.Context(), ctx.Writer())
		return
	}
	_, err = ctx.Req().Cookie(req.CookiePlayerId.Name)
	if err != nil {
		pagePlayerCreate(roomId).Render(ctx.Context(), ctx.Writer())
		return
	}

	pageRoomId(roomId).Render(ctx.Context(), ctx.Writer())
}

var wsUpgrader = websocket.Upgrader{}

func wsId(ctx maruchi.ReqContext) {
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
		// this does not remove a player if it is not the start of the game
		if game, ok := state.(*game_state.Game); ok {
			game.RemovePlayer(playerId.Value)
		}

		// todo: this deletes the room if 1 user refreshes, not really good experience
		// if len(players) == 0 {
		// 	reqContext.Rooms.DeleteRoom(roomId)
		// 	return nil
		// }

		if err := unsafeSyncGame(state, room.WsRoom()); err != nil {
			fmt.Println(err)
		}

		return nil
	})

	room.WsRoom().Add(ws, playerId.Value)
	defer room.WsRoom().Delete(ws) // 2. remove from ws room
	defer ws.Close()               // 1. close the ws conn

	if err := syncGame(room); err != nil {
		panic(err)
	}

	// the main game loop, events are as follows:
	// 1. read an action from a client
	// 2. change game state according to the action
	// 3. sync updated game state to all clients
	for {
		_, bytes, err := ws.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}

		var action game_state.Action
		if err := json.Unmarshal(bytes, &action); err != nil {
			panic(err)
		}

		switch action.Action {
		case game_state.ActionStartGame:
			if err := room.StateRef(func(state *game_state.GameState) error {
				game := (*state).(*game_state.Game)
				*state = game.Start()
				return nil
			}); err != nil {
				panic(err)
			}

		case game_state.ActionPlayerChooseWord:
			var action game_state.ActionPlayerChooseWordJson
			if err := json.Unmarshal(bytes, &action); err != nil {
				panic(err)
			}

			if err := room.StateRef(func(state *game_state.GameState) error {
				game := (*state).(*game_state.PlayerChooseWord)

				if !game_state.CheckSamePlayer(game, playerId.Value) {
					return fmt.Errorf("It's not your turn")
				}

				*state = game.Choose(action.WordIndex)
				return nil
			}); err != nil {
				panic(err)
			}
		case game_state.ActionPlayerSaySynonym:
			var action game_state.ActionPlayerSaySynonymJson
			if err := json.Unmarshal(bytes, &action); err != nil {
				panic(err)
			}

			if err := room.State(func(state game_state.GameState) error {
				game := state.(*game_state.PlayerTurn)

				if !game_state.CheckSamePlayer(game, playerId.Value) {
					return fmt.Errorf("It's not your turn")
				}

				game.SaySynonym(action.Synonym)
				return nil
			}); err != nil {
				panic(err)
			}
		default:
			panic("unsupported action")
		}

		if err := syncGame(room); err != nil {
			panic(err)
		}
	}
}

func unsafeSyncGame(state game_state.GameState, to *ws.Room) []error {
	return to.WriteEach(func(writer io.Writer, data any) error {
		return partialGameState(state, data.(string)).Render(context.Background(), writer)
	})
}

func syncGame(room *game_state.Room) error {
	return room.State(func(state game_state.GameState) error {
		if err := unsafeSyncGame(state, room.WsRoom()); err != nil {
			// todo:
			fmt.Println(err)
		}

		return nil
	})
}
