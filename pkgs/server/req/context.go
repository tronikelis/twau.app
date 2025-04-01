package req

import (
	"context"

	"word-amongus-game/pkgs/game_state"

	"github.com/tronikelis/maruchi"
)

func InitContext(request *maruchi.ReqContextBase, states game_state.States) {
	newContext := context.WithValue(
		request.Context(),
		game_state.States{},
		states,
	)

	request.R = request.R.WithContext(newContext)
}

func GetStates(ctx maruchi.ReqContext) game_state.States {
	return ctx.Context().Value(game_state.States{}).(game_state.States)
}
