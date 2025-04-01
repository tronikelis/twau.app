package rooms

import (
	"fmt"
	"net/http"

	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/server/req"

	"github.com/tronikelis/maruchi"
)

func postIndex(ctx maruchi.ReqContext) {
	playerName := ctx.Req().PostFormValue("player_name")

	playerId, err := game_state.RandomHex()
	if err != nil {
		panic(err)
	}

	gameId, err := game_state.RandomHex()
	if err != nil {
		panic(err)
	}

	game := req.GetStates(ctx).Upsert(gameId)

	game.AddPlayer(game_state.NewPlayer(playerId, playerName))

	ctx.Writer().Header().Set("hx-redirect", fmt.Sprintf("/rooms/%s", gameId))

	http.SetCookie(ctx.Writer(), &http.Cookie{
		Name:     "player_id",
		Value:    playerId,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
		MaxAge:   1 << 31,
	})
}
