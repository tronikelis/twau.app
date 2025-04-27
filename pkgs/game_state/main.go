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

type PlayerTurn interface {
	GameState
	Player() Player
}

type Expires interface {
	Expires() time.Time
}

type PlayerSynonym struct {
	Synonym  string
	PlayerId string
}

func newPlayerSynonym(synonym string, playerId string) PlayerSynonym {
	return PlayerSynonym{
		Synonym:  synonym,
		PlayerId: playerId,
	}
}

type Game struct {
	players   Players
	randomInt random.RandomIntNotSame
	word      string
	synonyms  []PlayerSynonym
	// ඞ
	imposterId string
}

func NewGame() *Game {
	return &Game{
		randomInt: random.NewRandomIntNotSame(3),
		players:   newPlayers(),
	}
}

func (self *Game) Players() Players {
	return self.players
}

func (self *Game) Word() string {
	return self.word
}

func (self *Game) ImposterId() string {
	return self.imposterId
}

// returns 0 value if missing
func (self *Game) Imposter() Player {
	player, _ := self.players.Player(self.imposterId)
	return player
}

func (self *Game) Synonyms() []PlayerSynonym {
	return self.synonyms
}

func (self *Game) GetGame() *Game {
	return self
}

func (self *Game) Start() *GamePlayerChooseWord {
	// clean garbanzo
	self.word = ""
	self.synonyms = nil
	self.players.ClearOffline()

	self.imposterId = self.players.Index(self.randomInt.IntN(self.players.Len())).Id

	return newGamePlayerChooseWord(self)
}

type playerVotePick struct {
	playerId string
	pickedId string
}

type GameVoteTurn struct {
	*Game
	picks        []playerVotePick
	playerId     string
	initPlayerId string
	candidates   Players
}

func newGameVoteTurn(game *Game, playerId string, candidates Players) *GameVoteTurn {
	return &GameVoteTurn{
		Game:         game,
		playerId:     playerId,
		initPlayerId: playerId,
		candidates:   candidates,
	}
}

func (self *GameVoteTurn) Player() Player {
	return self.players.PlayerOrPanic(self.playerId)
}

func (self *GameVoteTurn) InitPlayerId() string {
	return self.initPlayerId
}

// returns new game state
func (self *GameVoteTurn) Vote(playerId string) (GameState, bool) {
	self.picks = append(self.picks, playerVotePick{
		playerId: self.playerId,
		pickedId: playerId,
	})
	self.playerId = self.players.NextFrom(self.playerId).Id

	if self.playerId == self.initPlayerId {
		picks := map[string]int{}

		for _, v := range self.picks {
			picks[v.pickedId]++
		}

		highestCount := 0
		for _, v := range picks {
			if v > highestCount {
				highestCount = v
			}
		}

		var votedPlayerIds []string
		for k, v := range picks {
			if v == highestCount {
				votedPlayerIds = append(votedPlayerIds, k)
			}
		}

		// return tied players
		if len(votedPlayerIds) != 1 {
			// if imposter is not part of this list, means he won
			imposterInList := false
			for _, v := range votedPlayerIds {
				if self.Imposter().Id == v {
					imposterInList = true
				}
			}
			if !imposterInList {
				return newGameImposterWon(self.Game, self), true
			}

			candidates := newPlayers()
			for _, v := range votedPlayerIds {
				candidates.Add(self.players.PlayerOrPanic(v))
			}
			return newGameVoteTurn(self.Game, self.playerId, candidates), true
		}

		votedPlayerId := votedPlayerIds[0]

		voted := self.players.PlayerOrPanic(votedPlayerId)
		imposter := self.players.PlayerOrPanic(self.imposterId)

		// if imposter was picked, crewmates won
		if voted.Id == imposter.Id {
			return newGameCrewmateWon(self.Game, self), true
		}

		// imposter wasn't picked, he won
		return newGameImposterWon(self.Game, self), true
	}

	return nil, false
}

func (self *GameVoteTurn) Candidates(selfPlayerId string) []Player {
	players := self.candidates.Players()
	players = slices.DeleteFunc(players, func(a Player) bool {
		return a.Id == selfPlayerId
	})
	return players
}

type PlayerPicked struct {
	Player   Player
	PickedBy []Player
}

func (self *GameVoteTurn) Picks() []*PlayerPicked {
	candidates := self.candidates.Players()

	picks := make([]*PlayerPicked, len(candidates))
	picksMap := make(map[string]*PlayerPicked, len(candidates))

	for i, v := range picks {
		v = &PlayerPicked{}
		v.Player = candidates[i]

		picksMap[v.Player.Id] = v
		picks[i] = v
	}

	for _, v := range self.picks {
		prev := picksMap[v.pickedId]
		prev.PickedBy = append(
			prev.PickedBy,
			self.Game.players.PlayerOrPanic(v.playerId),
		)
	}

	slices.SortFunc(picks, func(a *PlayerPicked, b *PlayerPicked) int {
		return len(b.PickedBy) - len(a.PickedBy)
	})

	return picks
}

type GamePlayerTurn struct {
	*Game
	expires      time.Time
	playerId     string
	initPlayerId string
	fullCircle   bool
}

func newGamePlayerTurn(game *Game, playerId string) *GamePlayerTurn {
	return &GamePlayerTurn{
		Game:         game,
		playerId:     playerId,
		initPlayerId: playerId,
		expires:      time.Now().Add(PlayerTurnDuration),
	}
}

func (self *GamePlayerTurn) Expires() time.Time {
	return self.expires
}

func (self *GamePlayerTurn) FullCircle() bool {
	return self.fullCircle
}

func (self *GamePlayerTurn) InitVote() *GameVoteTurn {
	return newGameVoteTurn(self.Game, self.playerId, self.players)
}

func (self *GamePlayerTurn) Player() Player {
	return self.players.PlayerOrPanic(self.playerId)
}

// records player synonym and passes turn to next
// returns new game state
func (self *GamePlayerTurn) SaySynonym(synonym string) (GameState, bool) {
	self.synonyms = append(
		self.synonyms,
		newPlayerSynonym(synonym, self.playerId),
	)

	imposter := self.players.PlayerOrPanic(self.imposterId)
	current := self.players.PlayerOrPanic(self.playerId)

	// imposter could have won by saying the same word
	if synonym == self.word && imposter.Id == current.Id {
		return newGameImposterWon(self.Game, nil), true
	}

	self.playerId = self.players.NextFrom(self.playerId).Id
	if self.playerId == self.initPlayerId {
		self.fullCircle = true
	}

	self.expires = time.Now().Add(PlayerTurnDuration)

	return nil, false
}

type GamePlayerChooseWord struct {
	*Game
	playerId  string
	fromWords []string
}

func newGamePlayerChooseWord(game *Game) *GamePlayerChooseWord {
	player := game.players.Index(game.randomInt.IntN(game.players.Len()))

	// making imposter choose the word just does not make sense
	if player.Id == game.imposterId {
		player = game.players.NextFrom(player.Id)
	}

	return &GamePlayerChooseWord{
		Game:      game,
		fromWords: allWords.randomN(4),
		playerId:  player.Id,
	}
}

func (self *GamePlayerChooseWord) FromWords() []string {
	return self.fromWords
}

func (self *GamePlayerChooseWord) Player2() Player {
	return self.players.PlayerOrPanic(self.playerId)
}

func (self *GamePlayerChooseWord) Choose(index int) *GamePlayerTurn {
	if index < 0 || index >= len(self.fromWords) {
		index = 0
	}

	self.word = self.fromWords[index]
	player := self.players.Index(self.randomInt.IntN(self.players.Len()))
	return newGamePlayerTurn(self.Game, player.Id)
}

type GameCrewmateWon struct {
	*Game
	voteTurn *GameVoteTurn
}

func newGameCrewmateWon(game *Game, voteTurn *GameVoteTurn) *GameCrewmateWon {
	return &GameCrewmateWon{Game: game, voteTurn: voteTurn}
}

func (self *GameCrewmateWon) Picks() []*PlayerPicked {
	if self.voteTurn == nil {
		return nil
	}

	return self.voteTurn.Picks()
}

type GameImposterWon struct {
	*Game
	voteTurn *GameVoteTurn
}

func (self *GameImposterWon) Picks() []*PlayerPicked {
	if self.voteTurn == nil {
		return nil
	}

	return self.voteTurn.Picks()
}

func newGameImposterWon(game *Game, voteTurn *GameVoteTurn) *GameImposterWon {
	return &GameImposterWon{Game: game, voteTurn: voteTurn}
}
