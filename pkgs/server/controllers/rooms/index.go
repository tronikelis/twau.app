package rooms

import (
	"fmt"
	"net/http"

	"word-amongus-game/pkgs/auth"
	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/server/req"
)

func postIndex(ctx req.ReqContext) error {
	playerName := ctx.Req().PostFormValue("player_name")

	roomId, err := auth.RandomHex(auth.LengthRoomId)
	if err != nil {
		return err
	}

	state, ok := ctx.Rooms.CreateRoom(roomId)
	if !ok {
		return req.ErrRoomExists
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

	err = state.State(func(state game_state.GameState) error {
		game, ok := state.(*game_state.Game)
		if !ok {
			return fmt.Errorf("expected *game_state.Game, got %T", state)
		}

		game.AddPlayer(game_state.NewPlayer(playerCookies.Id.Value, playerCookies.Name.Value))

		return nil
	})
	if err != nil {
		return err
	}

	ctx.Writer().Header().Set("hx-redirect", fmt.Sprintf("/rooms/%s", roomId))

	return nil
}
