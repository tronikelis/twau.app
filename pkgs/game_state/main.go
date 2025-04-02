package game_state

import (
	"math/rand/v2"
	"slices"
)

type GameState interface {
	Game() *Game
}

type Player struct {
	Id   string
	Name string
}

func NewPlayer(id string, name string) Player {
	return Player{Id: id, Name: name}
}

type PlayerSynonym struct {
	synonym string
	player  Player
}

func newPlayerSynonym(synonym string, player Player) PlayerSynonym {
	return PlayerSynonym{
		synonym: synonym,
		player:  player,
	}
}

type Game struct {
	word     string
	synonyms []PlayerSynonym
	players  []Player
	// ඞ
	imposter Player
}

func NewGame() *Game {
	return &Game{}
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
}

func (self *Game) Start(word string) {
	self.word = word
	self.imposter = self.players[rand.IntN(len(self.players))]
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
	self.playerIndex = self.playerIndex + 1%len(self.game.players)

	if self.playerIndex == self.initPlayerIndex {
		return false
	}

	return true
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

// records player synonym and passes turn to next
func (self *PlayerTurn) SaySynonym(synonym string) {
	self.game.synonyms = append(
		self.game.synonyms,
		newPlayerSynonym(synonym, self.game.players[self.playerIndex]),
	)

	self.playerIndex = self.playerIndex + 1%len(self.game.players)
}
