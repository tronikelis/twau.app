package game_state

import "encoding/json"

const (
	ActionStartGame = iota
	ActionPlayerChooseWord
	ActionPlayerSaySynonym
	ActionInitVote
	ActionVote
)

type Action struct {
	Action int `json:"action"`
}

func SerializeJsonPanic(value any) string {
	str, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(str)
}

type ActionStartGameJson struct {
	Action
}

func NewActionStartGameJson() ActionStartGameJson {
	return ActionStartGameJson{
		Action: Action{ActionStartGame},
	}
}

type ActionPlayerChooseWordJson struct {
	Action
	WordIndex int `json:"word_index"`
}

func NewActionPlayerChooseWordJson(wordIndex int) ActionPlayerChooseWordJson {
	return ActionPlayerChooseWordJson{
		Action:    Action{ActionPlayerChooseWord},
		WordIndex: wordIndex,
	}
}

type ActionPlayerSaySynonymJson struct {
	Action
	// this will be populated with <input name="synonym"/>
	Synonym string `json:"synonym,omitempty"`
}

func NewActionPlayerSaySynonymJson() ActionPlayerSaySynonymJson {
	return ActionPlayerSaySynonymJson{
		Action: Action{ActionPlayerSaySynonym},
	}
}

type ActionInitVoteJson struct {
	Action
}

func NewActionInitVoteJson() ActionInitVoteJson {
	return ActionInitVoteJson{
		Action: Action{ActionInitVote},
	}
}

type ActionVoteJson struct {
	Action
	PlayerIndex int `json:"player_index"`
}

func NewActionVoteJson(playerIndex int) ActionVoteJson {
	return ActionVoteJson{
		Action:      Action{ActionVote},
		PlayerIndex: playerIndex,
	}
}
