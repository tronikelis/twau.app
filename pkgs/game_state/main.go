package game_state

import (
	"math/rand/v2"
	"slices"
)

const defaultTurnsLeft int = 10

type GameState interface {
	GetGame() *Game
}

type PlayerIndex interface {
	GameState
	PlayerIndex() int
}

func CheckSamePlayer(playerIndex PlayerIndex, playerId string) bool {
	return playerIndex.GetGame().Players()[playerIndex.PlayerIndex()].Id == playerId
}

type Player struct {
	Id   string
	Name string
}

func NewPlayer(id string, name string) Player {
	return Player{Id: id, Name: name}
}

type PlayerSynonym struct {
	Synonym string
	Player  Player
}

func newPlayerSynonym(synonym string, player Player) PlayerSynonym {
	return PlayerSynonym{
		Synonym: synonym,
		Player:  player,
	}
}

type Game struct {
	turnsLeft             int
	word                  string
	synonyms              []PlayerSynonym
	players               []Player
	prevChosenPlayerIndex int
	// ඞ
	imposter Player
}

func NewGame() *Game {
	return &Game{
		turnsLeft: defaultTurnsLeft,
	}
}

func (self *Game) Word() string {
	return self.word
}

func (self *Game) Imposter() Player {
	return self.imposter
}

func (self *Game) Synonyms() []PlayerSynonym {
	return self.synonyms
}

func (self *Game) RemovePlayer(id string) {
	if self.players == nil {
		return
	}

	self.players = slices.DeleteFunc(self.players, func(player Player) bool {
		return player.Id == id
	})
}

func (self *Game) HasPlayer(id string) bool {
	return slices.ContainsFunc(self.players, func(player Player) bool {
		return player.Id == id
	})
}

func (self *Game) Players() []Player {
	return self.players
}

func (self *Game) GetGame() *Game {
	return self
}

func (self *Game) Reset() {
	self.turnsLeft = defaultTurnsLeft
	self.word = ""
	self.synonyms = nil
	self.imposter = Player{}
	self.prevChosenPlayerIndex = -1
}

func (self *Game) Start() *PlayerChooseWord {
	imposterIndex := rand.IntN(len(self.players))
	self.imposter = self.players[imposterIndex]

	return NewPlayerChooseWord(self, imposterIndex)
}

// idempotent
func (self *Game) AddPlayer(player Player) {
	if self.HasPlayer(player.Id) {
		return
	}

	self.players = append(self.players, player)
}

func (self *Game) PlayerTurn(playerIndex int) *PlayerTurn {
	return newPlayerTurn(self, playerIndex)
}

type VoteTurn struct {
	*Game
	picks           map[Player]Player
	playerIndex     int
	initPlayerIndex int
}

func newVoteTurn(game *Game, playerIndex int) *VoteTurn {
	return &VoteTurn{
		Game:            game,
		picks:           map[Player]Player{},
		playerIndex:     playerIndex,
		initPlayerIndex: playerIndex,
	}
}

// returns false if voting has ended
func (self *VoteTurn) Vote(player Player) bool {
	self.picks[self.players[self.playerIndex]] = player
	self.playerIndex = (self.playerIndex + 1) % len(self.players)

	if self.playerIndex == self.initPlayerIndex {
		return false
	}

	return true
}

func (self *VoteTurn) Picks() map[Player]Player {
	return self.picks
}

type PlayerTurn struct {
	*Game
	playerIndex int
}

func newPlayerTurn(game *Game, playerIndex int) *PlayerTurn {
	return &PlayerTurn{
		Game:        game,
		playerIndex: playerIndex,
	}
}

func (self *PlayerTurn) InitVote() *VoteTurn {
	return newVoteTurn(self.Game, self.playerIndex)
}

func (self *PlayerTurn) PlayerIndex() int {
	return self.playerIndex
}

// records player synonym and passes turn to next
// returns (new game state, should set)
func (self *PlayerTurn) SaySynonym(synonym string) (GameState, bool) {
	self.synonyms = append(
		self.synonyms,
		newPlayerSynonym(synonym, self.players[self.playerIndex]),
	)

	if synonym == self.word && self.players[self.playerIndex].Id == self.imposter.Id {
		return NewImposterWon(self.Game), true
	}

	self.playerIndex = (self.playerIndex + 1) % len(self.players)

	return nil, false
}

type PlayerChooseWord struct {
	*Game
	playerIndex int
	fromWords   []string
}

func NewPlayerChooseWord(game *Game, imposterIndex int) *PlayerChooseWord {
	playerIndex := (game.prevChosenPlayerIndex + 1) % len(game.players)
	// making imposter choose the word just does not make sense
	if playerIndex == imposterIndex {
		playerIndex = (playerIndex + 1) % len(game.players)
	}

	return &PlayerChooseWord{
		Game:        game,
		fromWords:   allWords.RandomN(4),
		playerIndex: playerIndex,
	}
}

func (self *PlayerChooseWord) FromWords() []string {
	return self.fromWords
}

func (self *PlayerChooseWord) PlayerIndex() int {
	return self.playerIndex
}

func (self *PlayerChooseWord) Choose(index int) *PlayerTurn {
	if index < 0 || index >= len(self.fromWords) {
		index = 0
	}

	self.word = self.fromWords[index]
	return self.PlayerTurn(self.playerIndex)
}

type CrewmateWon struct {
	*Game
}

func NewCrewmateWon(game *Game) *CrewmateWon {
	return &CrewmateWon{Game: game}
}

type ImposterWon struct {
	*Game
}

func NewImposterWon(game *Game) *ImposterWon {
	return &ImposterWon{Game: game}
}
