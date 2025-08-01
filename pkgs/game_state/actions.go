package game_state

import "encoding/json"

const (
	ActionStart = iota
	ActionPlayerChooseWord
	ActionPlayerSaySynonym
	ActionInitVote
	ActionVote
	ActionPlayerChooseCategory
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

type ActionStartJson struct {
	Action
}

func NewActionStartJson() ActionStartJson {
	return ActionStartJson{
		Action: Action{ActionStart},
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
	PlayerId string `json:"player_id"`
}

func NewActionVoteJson(playerId string) ActionVoteJson {
	return ActionVoteJson{
		Action:   Action{ActionVote},
		PlayerId: playerId,
	}
}

type ActionPlayerChooseCategoryJson struct {
	Action
	CategoryId int `json:"category_id"`
}

func NewActionPlayerChooseCategoryJson(categoryId int) ActionPlayerChooseCategoryJson {
	return ActionPlayerChooseCategoryJson{
		Action:     Action{ActionPlayerChooseCategory},
		CategoryId: categoryId,
	}
}
