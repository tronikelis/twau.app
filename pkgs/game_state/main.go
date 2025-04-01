package game_state

import (
	"container/ring"
)

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
}

func (self *Game) AddPlayer(player Player) {
	r := &ring.Ring{Value: player}

	if self.players == nil {
		self.players = r
		return
	}

	self.players.Link(r)
}

func (self *Game) SetWord(word string) {
	self.word = word
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
