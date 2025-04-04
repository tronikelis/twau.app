package game_state

const (
	ACTION_START_GAME = "start_game"
)

type Action struct {
	Action string `json:"action"`
}
