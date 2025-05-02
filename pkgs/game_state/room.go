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
	password        string
}

func NewRoom(password string) *Room {
	room := &Room{
		wsRoom:          ws.NewRoom(),
		unsafeState:     NewGame(),
		mu:              &sync.Mutex{},
		stateChangeChan: make(chan bool),
		password:        password,
	}

	go room.asyncListenStateChange()

	return room
}

func (self *Room) Password() string {
	return self.password
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
		self.StateRef(func(state *GameState) error {
			*state = (*state).GetGame().Start()
			return nil
		})
	case ActionPlayerChooseWord:
		var action ActionPlayerChooseWordJson
		if err := json.Unmarshal(bytes, &action); err != nil {
			return err
		}

		if err := self.StateRef(func(state *GameState) error {
			game := (*state).(*GamePlayerChooseWord)

			if game.Player2().Id != playerId {
				return ErrNotYourTurn
			}

			*state = game.Choose(action.WordIndex)
			return nil
		}); err != nil {
			return err
		}
	case ActionPlayerSaySynonym:
		var action ActionPlayerSaySynonymJson
		if err := json.Unmarshal(bytes, &action); err != nil {
			return err
		}

		if err := self.StateRef(func(state *GameState) error {
			game := (*state).(*GamePlayerTurn)

			if game.Player().Id != playerId {
				return ErrNotYourTurn
			}

			if newState, ok := game.SaySynonym(action.Synonym); ok {
				*state = newState
			}
			return nil
		}); err != nil {
			return err
		}
	case ActionInitVote:
		if err := self.StateRef(func(state *GameState) error {
			game := (*state).(*GamePlayerTurn)

			if game.Player().Id != playerId {
				return ErrNotYourTurn
			}

			if !game.FullCircle() {
				return ErrBadAction
			}

			*state = game.InitVote()
			return nil
		}); err != nil {
			return err
		}
	case ActionVote:
		var action ActionVoteJson
		if err := json.Unmarshal(bytes, &action); err != nil {
			return err
		}

		if err := self.StateRef(func(state *GameState) error {
			game := (*state).(*GameVoteTurn)

			if game.Player().Id != playerId {
				return ErrNotYourTurn
			}

			if newState, ok := game.Vote(action.PlayerId); ok {
				*state = newState
			}
			return nil
		}); err != nil {
			return err
		}
	case ActionPlayerChooseCategory:
		var action ActionPlayerChooseCategoryJson
		if err := json.Unmarshal(bytes, &action); err != nil {
			return err
		}

		if err := self.StateRef(func(state *GameState) error {
			game := (*state).(*GamePlayerChooseCategory)

			if game.Player2().Id != playerId {
				return ErrNotYourTurn
			}

			newState, err := game.Choose(action.CategoryId)
			if err != nil {
				return err
			}

			*state = newState
			return nil
		}); err != nil {
			return err
		}
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
			self.stateRefNoChan(func(state *GameState) error {
				game, ok := (*state).(*GamePlayerTurn)
				if !ok {
					return nil
				}

				newState, ok := game.SaySynonym("")
				if !ok {
					return nil
				}

				*state = newState
				return nil
			})
		}
	}
}

func (self *Room) cleanup() {
	close(self.stateChangeChan)
}

func (self *Room) AddPlayer(conn *ws.ConnSafe, player Player) {
	self.wsRoom.Add(conn, player.Id)
	self.stateNoChan(func(state GameState) error {
		state.GetGame().Players().Add(player)
		return nil
	})
}

func (self *Room) RemovePlayer(conn *ws.ConnSafe, playerId string) {
	self.wsRoom.Delete(conn)
	self.stateNoChan(func(state GameState) error {
		if game, ok := state.(*Game); ok {
			game.Players().Remove(playerId)
		} else {
			state.GetGame().Players().Disconnect(playerId)
		}

		return nil
	})
}

func (self *Room) unsafeSyncGame() {
	if err := self.wsRoom.WriteEach(func(writer io.Writer, data any) error {
		return PartialGameState(self.unsafeState, data.(string)).Render(context.Background(), writer)
	}); err != nil {
		log.Println("unsafeSyncGame", "err", err)
	}
}

func (self *Room) stateNoChan(mutate func(state GameState) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	err := mutate(self.unsafeState)
	self.unsafeSyncGame()
	return err
}

func (self *Room) stateRefNoChan(mutate func(state *GameState) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	err := mutate(&self.unsafeState)
	self.unsafeSyncGame()
	return err
}

func (self *Room) StateRef(mutate func(state *GameState) error) error {
	err := self.stateRefNoChan(mutate)
	self.stateChangeChan <- true
	return err
}

func (self *Room) State(mutate func(state GameState) error) error {
	err := self.stateNoChan(mutate)
	self.stateChangeChan <- true
	return err
}
