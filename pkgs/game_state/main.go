package game_state

import (
	"math/rand/v2"
	"slices"
)

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
	word     string
	synonyms []PlayerSynonym
	players  []Player
	// ඞ
	imposter Player
}

func NewGame() *Game {
	return &Game{}
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
	self.word = ""
	self.synonyms = nil
	self.imposter = Player{}
}

func (self *Game) Start() *GamePlayerChooseWord {
	self.Reset()

	imposterIndex := rand.IntN(len(self.players))
	self.imposter = self.players[imposterIndex]

	return newGamePlayerChooseWord(self, imposterIndex)
}

// idempotent
func (self *Game) AddPlayer(player Player) {
	if self.HasPlayer(player.Id) {
		return
	}

	self.players = append(self.players, player)
}

func (self *Game) PlayerTurn(playerIndex int) *GamePlayerTurn {
	return newGamePlayerTurn(self, playerIndex)
}

type GameVoteTurn struct {
	*Game
	picks           map[Player]Player
	playerIndex     int
	initPlayerIndex int
}

func newGameVoteTurn(game *Game, playerIndex int) *GameVoteTurn {
	return &GameVoteTurn{
		Game:            game,
		picks:           map[Player]Player{},
		playerIndex:     playerIndex,
		initPlayerIndex: playerIndex,
	}
}

// returns false if voting has ended
func (self *GameVoteTurn) Vote(player Player) bool {
	self.picks[self.players[self.playerIndex]] = player
	self.playerIndex = (self.playerIndex + 1) % len(self.players)

	if self.playerIndex == self.initPlayerIndex {
		return false
	}

	return true
}

func (self *GameVoteTurn) Picks() map[Player]Player {
	return self.picks
}

type GamePlayerTurn struct {
	*Game
	playerIndex int
}

func newGamePlayerTurn(game *Game, playerIndex int) *GamePlayerTurn {
	return &GamePlayerTurn{
		Game:        game,
		playerIndex: playerIndex,
	}
}

func (self *GamePlayerTurn) InitVote() *GameVoteTurn {
	return newGameVoteTurn(self.Game, self.playerIndex)
}

func (self *GamePlayerTurn) PlayerIndex() int {
	return self.playerIndex
}

// records player synonym and passes turn to next
// returns (new game state, should set)
func (self *GamePlayerTurn) SaySynonym(synonym string) (GameState, bool) {
	self.synonyms = append(
		self.synonyms,
		newPlayerSynonym(synonym, self.players[self.playerIndex]),
	)

	if synonym == self.word && self.players[self.playerIndex].Id == self.imposter.Id {
		return newGameImposterWon(self.Game), true
	}

	self.playerIndex = (self.playerIndex + 1) % len(self.players)

	return nil, false
}

type GamePlayerChooseWord struct {
	*Game
	playerIndex int
	fromWords   []string
}

func newGamePlayerChooseWord(game *Game, imposterIndex int) *GamePlayerChooseWord {
	playerIndex := rand.IntN(len(game.players))
	// making imposter choose the word just does not make sense
	if playerIndex == imposterIndex {
		playerIndex = (playerIndex + 1) % len(game.players)
	}

	return &GamePlayerChooseWord{
		Game:        game,
		fromWords:   allWords.RandomN(4),
		playerIndex: playerIndex,
	}
}

func (self *GamePlayerChooseWord) FromWords() []string {
	return self.fromWords
}

func (self *GamePlayerChooseWord) PlayerIndex() int {
	return self.playerIndex
}

func (self *GamePlayerChooseWord) Choose(index int) *GamePlayerTurn {
	if index < 0 || index >= len(self.fromWords) {
		index = 0
	}

	self.word = self.fromWords[index]
	return self.PlayerTurn(self.playerIndex)
}

type GameCrewmateWon struct {
	*Game
}

func newGameCrewmateWon(game *Game) *GameCrewmateWon {
	return &GameCrewmateWon{Game: game}
}

type GameImposterWon struct {
	*Game
}

func newGameImposterWon(game *Game) *GameImposterWon {
	return &GameImposterWon{Game: game}
}
