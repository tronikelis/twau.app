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

func CheckSamePlayer(game GameState, playerIndex int, playerId string) bool {
	return game.GetGame().Players()[playerIndex].Id == playerId
}

type PlayerWithIndex struct {
	Player
	Index int
}

type Player struct {
	Id     string
	Name   string
	Online bool
}

func NewPlayer(id string, name string) Player {
	return Player{Id: id, Name: name, Online: true}
}

type PlayerSynonym struct {
	Synonym     string
	PlayerIndex int
}

func newPlayerSynonym(synonym string, playerIndex int) PlayerSynonym {
	return PlayerSynonym{
		Synonym:     synonym,
		PlayerIndex: playerIndex,
	}
}

type Game struct {
	word     string
	synonyms []PlayerSynonym
	players  []Player
	// ඞ
	imposterIndex int
}

func NewGame() *Game {
	return &Game{
		imposterIndex: -1,
	}
}

func (self *Game) Word() string {
	return self.word
}

func (self *Game) Imposter() Player {
	if self.imposterIndex == -1 {
		return Player{}
	}

	return self.players[self.imposterIndex]
}

func (self *Game) Synonyms() []PlayerSynonym {
	return self.synonyms
}

func (self *Game) DisconnectPlayer(id string) {
	for i, v := range self.players {
		if v.Id == id {
			v.Online = false
			self.players[i] = v
			break
		}
	}
}

func (self *Game) PlayersOnline() int {
	count := 0
	for _, v := range self.players {
		if v.Online {
			count++
		}
	}
	return count
}

func (self *Game) RemovePlayer(id string) {
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

func (self *Game) Start() *GamePlayerChooseWord {
	// clean garbanzo
	self.word = ""
	self.synonyms = nil
	self.players = slices.DeleteFunc(self.players, func(v Player) bool {
		return !v.Online
	})

	self.imposterIndex = rand.IntN(len(self.players))

	return newGamePlayerChooseWord(self, self.imposterIndex)
}

// idempotent
func (self *Game) AddPlayer(player Player) {
	if self.HasPlayer(player.Id) {
		return
	}

	self.players = append(self.players, player)
}

type playerVotePick struct {
	playerIndex int
	pickedIndex int
}

type GameVoteTurn struct {
	*Game
	picks           []playerVotePick
	playerIndex     int
	initPlayerIndex int
}

func newGameVoteTurn(game *Game, playerIndex int) *GameVoteTurn {
	return &GameVoteTurn{
		Game:            game,
		playerIndex:     playerIndex,
		initPlayerIndex: playerIndex,
	}
}

func (self *GameVoteTurn) PlayerIndex() int {
	return self.playerIndex
}

// returns new game state
func (self *GameVoteTurn) Vote(playerIndex int) (GameState, bool) {
	self.picks = append(self.picks, playerVotePick{
		playerIndex: self.playerIndex,
		pickedIndex: playerIndex,
	})
	self.playerIndex = (self.playerIndex + 1) % len(self.players)

	if self.playerIndex == self.initPlayerIndex {
		picks := map[int]int{}

		for _, v := range self.picks {
			picks[v.pickedIndex]++
		}

		highestCount := 0
		pickedPlayerIndex := 0

		for k, v := range picks {
			if v > highestCount {
				highestCount = v
				pickedPlayerIndex = k
			}
		}

		// if imposter was picked, crewmates won
		if self.players[pickedPlayerIndex].Id == self.players[self.imposterIndex].Id {
			return newGameCrewmateWon(self.Game), true
		}

		// imposter wasn't picked, he won
		return newGameImposterWon(self.Game), true
	}

	return nil, false
}

func (self *GameVoteTurn) Players(selfPlayerId string) []PlayerWithIndex {
	players := make([]PlayerWithIndex, 0, len(self.players)-1)
	for i, v := range self.players {
		// skip adding self as you can't vote yourself out
		if v.Id == selfPlayerId {
			continue
		}

		players = append(players, PlayerWithIndex{
			Player: v,
			Index:  i,
		})
	}
	return players
}

type PlayerPicked struct {
	Player   Player
	PickedBy []Player
}

func (self *GameVoteTurn) Picks() []PlayerPicked {
	picks := make([]PlayerPicked, len(self.players))
	for i, v := range picks {
		v.Player = self.players[i]
		picks[i] = v
	}

	for _, v := range self.picks {
		prev := picks[v.pickedIndex]
		prev.PickedBy = append(prev.PickedBy, self.players[v.playerIndex])
		picks[v.pickedIndex] = prev
	}

	slices.SortFunc(picks, func(a PlayerPicked, b PlayerPicked) int {
		return len(b.PickedBy) - len(a.PickedBy)
	})

	return picks
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
// returns new game state
func (self *GamePlayerTurn) SaySynonym(synonym string) (GameState, bool) {
	self.synonyms = append(
		self.synonyms,
		newPlayerSynonym(synonym, self.playerIndex),
	)

	// imposter could have won by saying the same word
	if synonym == self.word && self.players[self.playerIndex].Id == self.players[self.imposterIndex].Id {
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
		fromWords:   allWords.randomN(4),
		playerIndex: playerIndex,
	}
}

func (self *GamePlayerChooseWord) FromWords() []string {
	return self.fromWords
}

func (self *GamePlayerChooseWord) PlayerIndex2() int {
	return self.playerIndex
}

func (self *GamePlayerChooseWord) Choose(index int) *GamePlayerTurn {
	if index < 0 || index >= len(self.fromWords) {
		index = 0
	}

	self.word = self.fromWords[index]
	playerIndex := rand.IntN(len(self.players))
	return newGamePlayerTurn(self.Game, playerIndex)
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
