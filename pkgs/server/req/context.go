package req

import (
	"context"

	"word-amongus-game/pkgs/game_state"

	"github.com/tronikelis/maruchi"
)

const (
	statesKey int = iota
)

func InitContext(request *maruchi.ReqContextBase, states game_state.States) {
	newContext := context.WithValue(
		request.Context(),
		statesKey,
		states,
	)

	request.R = request.R.WithContext(newContext)
}

func GetStates(ctx maruchi.ReqContext) game_state.States {
	return ctx.Context().Value(statesKey).(game_state.States)
}
