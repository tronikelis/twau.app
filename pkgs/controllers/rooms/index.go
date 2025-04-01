package rooms

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"word-amongus-game/pkgs/game_state"

	"github.com/tronikelis/maruchi"
)

func randomHex() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

var games map[string]any = map[string]any{}

func postIndex(ctx maruchi.ReqContext) {
	playerName := ctx.Req().PostFormValue("player_name")

	playerId, err := randomHex()
	if err != nil {
		panic(err)
	}

	gameId, err := randomHex()
	if err != nil {
		panic(err)
	}

	game := game_state.NewGame()
	game.AddPlayer(game_state.NewPlayer(playerId, playerName))

	games[gameId] = game

	ctx.Writer().Header().Set("hx-redirect", fmt.Sprintf("/rooms/%s", gameId))

	http.SetCookie(ctx.Writer(), &http.Cookie{
		Name:     "player_id",
		Value:    playerId,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
		MaxAge:   1 << 31,
	})
}
