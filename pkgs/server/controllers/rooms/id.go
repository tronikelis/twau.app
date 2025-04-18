package rooms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/server/req"
	"word-amongus-game/pkgs/ws"

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

	log.Println(fmt.Sprintf("%s connected, [%s]", playerCookies.Name.Value, playerCookies.Id.Value))

	if err := room.State(func(state game_state.GameState) error {
		state.GetGame().AddPlayer(game_state.NewPlayer(playerCookies.Id.Value, playerCookies.Name.Value))
		return nil
	}); err != nil {
		return err
	}
	defer room.State(func(state game_state.GameState) error { // 3. sync changes to others
		if game, ok := state.(*game_state.Game); ok {
			game.RemovePlayer(playerCookies.Id.Value)
		} else {
			state.GetGame().DisconnectPlayer(playerCookies.Id.Value)
		}

		if state.GetGame().PlayersOnline() == 0 {
			ctx.Rooms.DeleteRoom(roomId)
			log.Println("deleting room", roomId)
			return nil
		}

		if err := unsafeSyncGame(state, room.WsRoom()); err != nil {
			log.Println("unsafeSyncGame", "err", err)
		}

		return nil
	})

	room.WsRoom().Add(socket, playerCookies.Id.Value)
	defer room.WsRoom().Delete(socket) // 2. remove from ws room
	defer socket.Close()               // 1. close the ws conn

	if err := syncGame(room); err != nil {
		log.Println("syncGame", "err", err)
	}

	// the main game loop, events are as follows:
	// 1. read an action from a client
	// 2. change game state according to the action
	// 3. sync updated game state to all clients
	for {
		_, bytes, err := socket.ReadMessage()
		if err != nil {
			return err
		}

		var action game_state.Action
		if err := json.Unmarshal(bytes, &action); err != nil {
			return err
		}

		switch action.Action {
		case game_state.ActionStart:
			if err := room.StateRef(func(state *game_state.GameState) error {
				game := (*state).(*game_state.Game)
				*state = game.Start()
				return nil
			}); err != nil {
				return err
			}

		case game_state.ActionPlayerChooseWord:
			var action game_state.ActionPlayerChooseWordJson
			if err := json.Unmarshal(bytes, &action); err != nil {
				return err
			}

			if err := room.StateRef(func(state *game_state.GameState) error {
				game := (*state).(*game_state.GamePlayerChooseWord)

				if !game_state.CheckSamePlayer(game, playerCookies.Id.Value) {
					return req.ErrNotYourTurn
				}

				*state = game.Choose(action.WordIndex)
				return nil
			}); err != nil {
				return err
			}
		case game_state.ActionPlayerSaySynonym:
			var action game_state.ActionPlayerSaySynonymJson
			if err := json.Unmarshal(bytes, &action); err != nil {
				return err
			}

			if err := room.StateRef(func(state *game_state.GameState) error {
				game := (*state).(*game_state.GamePlayerTurn)

				if !game_state.CheckSamePlayer(game, playerCookies.Id.Value) {
					return req.ErrNotYourTurn
				}

				if newState, ok := game.SaySynonym(action.Synonym); ok {
					*state = newState
				}

				return nil
			}); err != nil {
				return err
			}
		case game_state.ActionInitVote:
			if err := room.StateRef(func(state *game_state.GameState) error {
				game := (*state).(*game_state.GamePlayerTurn)

				if !game_state.CheckSamePlayer(game, playerCookies.Id.Value) {
					return req.ErrNotYourTurn
				}

				*state = game.InitVote()
				return nil
			}); err != nil {
				return err
			}
		case game_state.ActionVote:
			var action game_state.ActionVoteJson
			if err := json.Unmarshal(bytes, &action); err != nil {
				return err
			}

			if err := room.StateRef(func(state *game_state.GameState) error {
				game := (*state).(*game_state.GameVoteTurn)

				if !game_state.CheckSamePlayer(game, playerCookies.Id.Value) {
					return req.ErrNotYourTurn
				}

				if newState, ok := game.Vote(action.PlayerIndex); ok {
					*state = newState
				}
				return nil
			}); err != nil {
				return err
			}
		case game_state.ActionRestart:
			if err := room.StateRef(func(state *game_state.GameState) error {
				game := (*state).GetGame()
				*state = game.Start()
				return nil
			}); err != nil {
				return err
			}
		default:
			return req.ErrUnknownAction
		}

		if err := syncGame(room); err != nil {
			log.Println("syncGame", "err", err)
		}
	}
}

func unsafeSyncGame(state game_state.GameState, to *ws.Room) error {
	return to.WriteEach(func(writer io.Writer, data any) error {
		return partialGameState(state, data.(string)).Render(context.Background(), writer)
	})
}

func syncGame(room *game_state.Room) error {
	return room.State(func(state game_state.GameState) error {
		return unsafeSyncGame(state, room.WsRoom())
	})
}
