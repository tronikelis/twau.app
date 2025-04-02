package rooms

import (
	"fmt"
	"net/http"

	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/server/req"

	"github.com/tronikelis/maruchi"
)

func postIndex(ctx maruchi.ReqContext) {
	reqContext := req.GetReqContext(ctx)

	playerName := ctx.Req().PostFormValue("player_name")

	playerId, err := game_state.RandomHex()
	if err != nil {
		panic(err)
	}

	roomId, err := game_state.RandomHex()
	if err != nil {
		panic(err)
	}

	state, ok := reqContext.Rooms.CreateRoom(roomId)
	if !ok {
		panic("room already exists")
	}

	err = state.State(func(state game_state.GameState) error {
		game, ok := state.(*game_state.Game)
		if !ok {
			return fmt.Errorf("expected *game_state.Game, got %T", state)
		}

		game.AddPlayer(game_state.NewPlayer(playerId, playerName))

		return nil
	})
	if err != nil {
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
