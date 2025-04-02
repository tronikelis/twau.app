package game_state

import (
	"container/ring"
	"math/rand/v2"
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

func NewGame() *Game {
	return &Game{}
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
	players  *ring.Ring
	// ඞ
	imposter Player
}

// return false to stop iteration
func (self *Game) eachPlayer(do func(player *ring.Ring) bool) {
	if self.players == nil {
		return
	}

	if self.players.Len() == 1 {
		do(self.players)
		return
	}

	curr := self.players
	// we have a do while at home honey
	for ok := true; ok; ok = curr != self.players {
		if !do(curr) {
			return
		}

		curr = curr.Next()
	}
}

func (self *Game) RemovePlayer(id string) {
	if self.players == nil {
		return
	}

	// special case, 1 element v.Next() == itself
	if self.players.Len() == 1 {
		self.players = nil
		return
	}

	self.eachPlayer(func(player *ring.Ring) bool {
		if player.Value.(Player).Id == id {
			self.players.Prev().Unlink(1)
			return false
		}

		return true
	})
}

func (self *Game) HasPlayer(id string) bool {
	has := false
	self.eachPlayer(func(player *ring.Ring) bool {
		if player.Value.(Player).Id == id {
			has = true
			return false
		}

		return true
	})

	return has
}

func (self *Game) Players() []Player {
	if self.players == nil {
		return nil
	}

	var players []Player

	self.players.Do(func(a any) {
		players = append(players, a.(Player))
	})

	return players
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

	imposterIndex := rand.Uint32N(uint32(self.players.Len()))
	imposter := self.players

	for range imposterIndex {
		imposter = imposter.Next()
	}

	self.imposter = imposter.Value.(Player)
}

// idempotent
func (self *Game) AddPlayer(player Player) {
	if self.HasPlayer(player.Id) {
		return
	}

	r := ring.New(1)
	r.Value = player

	if self.players == nil {
		self.players = r
		return
	}

	self.players.Link(r)
}

func (self *Game) PlayerTurn() *PlayerTurn {
	return newPlayerTurn(self, self.players)
}

type VoteTurn struct {
	game *Game

	picks      map[Player]Player
	initPlayer *ring.Ring
	player     *ring.Ring
}

func (self *VoteTurn) Game() *Game {
	return self.game
}

func newVoteTurn(game *Game, player *ring.Ring) *VoteTurn {
	return &VoteTurn{
		game:       game,
		picks:      map[Player]Player{},
		player:     player,
		initPlayer: player,
	}
}

// returns false if voting has ended
func (self *VoteTurn) Vote(player Player) bool {
	self.picks[self.player.Value.(Player)] = player
	self.player = self.player.Next()

	if self.player == self.initPlayer {
		return false
	}

	return true
}

type PlayerTurn struct {
	game   *Game
	player *ring.Ring
}

func (self *PlayerTurn) Game() *Game {
	return self.game
}

func newPlayerTurn(game *Game, player *ring.Ring) *PlayerTurn {
	return &PlayerTurn{
		game:   game,
		player: player,
	}
}

func (self *PlayerTurn) InitVote() *VoteTurn {
	return newVoteTurn(self.game, self.player)
}

// records player synonym and passes turn to next
func (self *PlayerTurn) SaySynonym(synonym string) {
	self.game.synonyms = append(
		self.game.synonyms,
		newPlayerSynonym(synonym, self.player.Value.(Player)),
	)

	self.player = self.player.Next()
}
