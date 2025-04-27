package game_state

import (
	"slices"
	"time"

	"twau.app/pkgs/random"
)

const PlayerTurnDuration = time.Minute * 2

type GameState interface {
	GetGame() *Game
}

type PlayerIndex interface {
	GameState
	PlayerIndex() int
}

type PlayerWithIndex struct {
	Player
	Index int
}

type Expires interface {
	Expires() time.Time
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
	randomInt random.RandomIntNotSame
	word      string
	synonyms  []PlayerSynonym
	players   []Player
	// ඞ
	imposterIndex int
}

func NewGame() *Game {
	return &Game{
		imposterIndex: -1,
		randomInt:     random.NewRandomIntNotSame(3),
	}
}

func (self *Game) Word() string {
	return self.word
}

func (self *Game) ImposterIndex() int {
	return self.imposterIndex
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

	self.imposterIndex = self.randomInt.IntN(len(self.players))

	return newGamePlayerChooseWord(self, self.imposterIndex)
}

// idempotent
func (self *Game) AddPlayer(player Player) {
	index := slices.IndexFunc(self.players, func(v Player) bool {
		return v.Id == player.Id
	})

	if index != -1 {
		self.players[index] = player
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

func (self *GameVoteTurn) InitPlayerIndex() int {
	return self.initPlayerIndex
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

		dupHighestCount := 0
		for _, v := range picks {
			if v == highestCount {
				dupHighestCount++
			}
		}

		// means multiple players have the highest vote count
		// in other words a tie
		if dupHighestCount != 1 {
			return newGameVoteTurn(self.Game, self.initPlayerIndex), true
		}

		// if imposter was picked, crewmates won
		if self.players[pickedPlayerIndex].Id == self.players[self.imposterIndex].Id {
			return newGameCrewmateWon(self.Game, self), true
		}

		// imposter wasn't picked, he won
		return newGameImposterWon(self.Game, self), true
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
	Player   PlayerWithIndex
	PickedBy []PlayerWithIndex
}

func (self *GameVoteTurn) Picks() []PlayerPicked {
	picks := make([]PlayerPicked, len(self.players))
	for i, v := range picks {
		v.Player = PlayerWithIndex{
			Player: self.players[i],
			Index:  i,
		}
		picks[i] = v
	}

	for _, v := range self.picks {
		prev := picks[v.pickedIndex]
		prev.PickedBy = append(
			prev.PickedBy,
			PlayerWithIndex{
				Player: self.players[v.playerIndex],
				Index:  v.playerIndex,
			},
		)
		picks[v.pickedIndex] = prev
	}

	slices.SortFunc(picks, func(a PlayerPicked, b PlayerPicked) int {
		return len(b.PickedBy) - len(a.PickedBy)
	})

	return picks
}

type GamePlayerTurn struct {
	*Game
	expires         time.Time
	playerIndex     int
	initPlayerIndex int
	fullCircle      bool
}

func newGamePlayerTurn(game *Game, playerIndex int) *GamePlayerTurn {
	return &GamePlayerTurn{
		Game:            game,
		playerIndex:     playerIndex,
		initPlayerIndex: playerIndex,
		expires:         time.Now().Add(PlayerTurnDuration),
	}
}

func (self *GamePlayerTurn) Expires() time.Time {
	return self.expires
}

func (self *GamePlayerTurn) FullCircle() bool {
	return self.fullCircle
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
		return newGameImposterWon(self.Game, nil), true
	}

	self.playerIndex = (self.playerIndex + 1) % len(self.players)

	if self.playerIndex == self.initPlayerIndex {
		self.fullCircle = true
	}

	self.expires = time.Now().Add(PlayerTurnDuration)

	return nil, false
}

type GamePlayerChooseWord struct {
	*Game
	playerIndex int
	fromWords   []string
}

func newGamePlayerChooseWord(game *Game, imposterIndex int) *GamePlayerChooseWord {
	playerIndex := game.randomInt.IntN(len(game.players))
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
	playerIndex := self.randomInt.IntN(len(self.players))
	return newGamePlayerTurn(self.Game, playerIndex)
}

type GameCrewmateWon struct {
	*Game
	voteTurn *GameVoteTurn
}

func newGameCrewmateWon(game *Game, voteTurn *GameVoteTurn) *GameCrewmateWon {
	return &GameCrewmateWon{Game: game, voteTurn: voteTurn}
}

func (self *GameCrewmateWon) Picks() []PlayerPicked {
	if self.voteTurn == nil {
		return nil
	}

	return self.voteTurn.Picks()
}

type GameImposterWon struct {
	*Game
	voteTurn *GameVoteTurn
}

func (self *GameImposterWon) Picks() []PlayerPicked {
	if self.voteTurn == nil {
		return nil
	}

	return self.voteTurn.Picks()
}

func newGameImposterWon(game *Game, voteTurn *GameVoteTurn) *GameImposterWon {
	return &GameImposterWon{Game: game, voteTurn: voteTurn}
}
