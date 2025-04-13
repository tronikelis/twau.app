package game_state

import (
	"math/rand/v2"
	"slices"
)

type GameState interface {
	Game() *Game
}

type PlayerIndex interface {
	GameState
	PlayerIndex() int
}

func CheckSamePlayer(playerIndex PlayerIndex, playerId string) bool {
	return playerIndex.Game().Players()[playerIndex.PlayerIndex()].Id == playerId
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
	word                  string
	synonyms              []PlayerSynonym
	players               []Player
	prevChosenPlayerIndex int
	// ඞ
	imposter Player
}

func NewGame() *Game {
	return &Game{}
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

func (self *Game) Game() *Game {
	return self
}

func (self *Game) Reset() {
	self.word = ""
	self.synonyms = nil
	self.imposter = Player{}
	self.prevChosenPlayerIndex = 0
}

func (self *Game) Start() *PlayerChooseWord {
	self.imposter = self.players[rand.IntN(len(self.players))]
	return NewPlayerChooseWord(self)
}

// idempotent
func (self *Game) AddPlayer(player Player) {
	if self.HasPlayer(player.Id) {
		return
	}

	self.players = append(self.players, player)
}

func (self *Game) PlayerTurn() *PlayerTurn {
	return newPlayerTurn(self)
}

type VoteTurn struct {
	game *Game

	picks           map[Player]Player
	playerIndex     int
	initPlayerIndex int
}

func newVoteTurn(game *Game, playerIndex int) *VoteTurn {
	return &VoteTurn{
		game:            game,
		picks:           map[Player]Player{},
		playerIndex:     playerIndex,
		initPlayerIndex: playerIndex,
	}
}

func (self *VoteTurn) Game() *Game {
	return self.game
}

// returns false if voting has ended
func (self *VoteTurn) Vote(player Player) bool {
	self.picks[self.game.players[self.playerIndex]] = player
	self.playerIndex = (self.playerIndex + 1) % len(self.game.players)

	if self.playerIndex == self.initPlayerIndex {
		return false
	}

	return true
}

func (self *VoteTurn) Picks() map[Player]Player {
	return self.picks
}

type PlayerTurn struct {
	game        *Game
	playerIndex int
}

func newPlayerTurn(game *Game) *PlayerTurn {
	return &PlayerTurn{
		game: game,
	}
}

func (self *PlayerTurn) Game() *Game {
	return self.game
}

func (self *PlayerTurn) InitVote() *VoteTurn {
	return newVoteTurn(self.game, self.playerIndex)
}

func (self *PlayerTurn) PlayerIndex() int {
	return self.playerIndex
}

// records player synonym and passes turn to next
func (self *PlayerTurn) SaySynonym(synonym string) {
	self.game.synonyms = append(
		self.game.synonyms,
		newPlayerSynonym(synonym, self.game.players[self.playerIndex]),
	)

	self.playerIndex = (self.playerIndex + 1) % len(self.game.players)
}

type PlayerChooseWord struct {
	game        *Game
	playerIndex int
	fromWords   []string
}

func NewPlayerChooseWord(game *Game) *PlayerChooseWord {
	return &PlayerChooseWord{
		game:        game,
		fromWords:   allWords.RandomN(4),
		playerIndex: (game.prevChosenPlayerIndex + 1) % len(game.players),
	}
}

func (self *PlayerChooseWord) Game() *Game {
	return self.game
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

	self.game.word = self.fromWords[index]
	return self.game.PlayerTurn()
}
