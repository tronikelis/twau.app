package game_state

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"twau.app/pkgs/ws"
)

type Room struct {
	wsRoom          *ws.Room
	unsafeState     GameState
	mu              *sync.Mutex
	stateChangeChan chan bool
}

func NewRoom() *Room {
	room := &Room{
		wsRoom:          ws.NewRoom(),
		unsafeState:     NewGame(),
		mu:              &sync.Mutex{},
		stateChangeChan: make(chan bool),
	}

	go room.asyncListenStateChange()

	return room
}

// the main game loop, events are as follows:
// 1. read an action from a client
// 2. change game state according to the action
// 3. sync updated game state to all clients
func (self *Room) GameLoop(bytes []byte, playerId string) error {
	var action Action
	if err := json.Unmarshal(bytes, &action); err != nil {
		return err
	}

	switch action.Action {
	case ActionStart:
		self.StateRef(func(state *GameState) {
			*state = (*state).GetGame().Start()
		})
	case ActionPlayerChooseWord:
		var action ActionPlayerChooseWordJson
		if err := json.Unmarshal(bytes, &action); err != nil {
			return err
		}

		self.StateRef(func(state *GameState) {
			game := (*state).(*GamePlayerChooseWord)

			if game.PlayerId2() != playerId {
				return
			}

			*state = game.Choose(action.WordIndex)
		})
	case ActionPlayerSaySynonym:
		var action ActionPlayerSaySynonymJson
		if err := json.Unmarshal(bytes, &action); err != nil {
			return err
		}

		self.StateRef(func(state *GameState) {
			game := (*state).(*GamePlayerTurn)

			if game.PlayerId() != playerId {
				return
			}

			if newState, ok := game.SaySynonym(action.Synonym); ok {
				*state = newState
			}
		})
	case ActionInitVote:
		self.StateRef(func(state *GameState) {
			game := (*state).(*GamePlayerTurn)

			if game.PlayerId() != playerId {
				return
			}

			if !game.FullCircle() {
				return
			}

			*state = game.InitVote()
		})
	case ActionVote:
		var action ActionVoteJson
		if err := json.Unmarshal(bytes, &action); err != nil {
			return err
		}

		self.StateRef(func(state *GameState) {
			game := (*state).(*GameVoteTurn)

			if game.PlayerId() != playerId {
				return
			}

			if newState, ok := game.Vote(action.PlayerId); ok {
				*state = newState
			}
		})
	default:
		return ErrUnknownAction
	}

	return nil
}

func (self *Room) asyncListenStateChange() {
	for {
		select {
		case ok := <-self.stateChangeChan:
			if !ok {
				return
			}
		case <-time.After(PlayerTurnDuration):
			self.stateRefNoChan(func(state *GameState) {
				game, ok := (*state).(*GamePlayerTurn)
				if !ok {
					return
				}

				newState, ok := game.SaySynonym("")
				if !ok {
					return
				}

				*state = newState
			})
		}
	}
}

func (self *Room) cleanup() {
	close(self.stateChangeChan)
}

func (self *Room) AddPlayer(conn *ws.ConnSafe, player Player) {
	self.wsRoom.Add(conn, player.Id)
	self.stateNoChan(func(state GameState) {
		state.GetGame().Players().Add(player)
	})
}

func (self *Room) RemovePlayer(conn *ws.ConnSafe, playerId string) {
	self.wsRoom.Delete(conn)
	self.stateNoChan(func(state GameState) {
		if game, ok := state.(*Game); ok {
			game.Players().Remove(playerId)
		} else {
			state.GetGame().Players().Disconnect(playerId)
		}
	})
}

func (self *Room) unsafeSyncGame() {
	if err := self.wsRoom.WriteEach(func(writer io.Writer, data any) error {
		return PartialGameState(self.unsafeState, data.(string)).Render(context.Background(), writer)
	}); err != nil {
		log.Println("unsafeSyncGame", "err", err)
	}
}

func (self *Room) stateNoChan(mutate func(state GameState)) {
	self.mu.Lock()
	defer self.mu.Unlock()

	mutate(self.unsafeState)
	self.unsafeSyncGame()
}

func (self *Room) stateRefNoChan(mutate func(state *GameState)) {
	self.mu.Lock()
	defer self.mu.Unlock()

	mutate(&self.unsafeState)
	self.unsafeSyncGame()
}

func (self *Room) StateRef(mutate func(state *GameState)) {
	self.stateRefNoChan(mutate)
	self.stateChangeChan <- true
}

func (self *Room) State(mutate func(state GameState)) {
	self.stateNoChan(mutate)
	self.stateChangeChan <- true
}
