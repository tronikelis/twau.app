package game_state

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"word-amongus-game/pkgs/ws"

	"github.com/gorilla/websocket"
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
func (self *Room) GameLoop(conn *websocket.Conn, playerId string) error {
	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var action Action
		if err := json.Unmarshal(bytes, &action); err != nil {
			return err
		}

		switch action.Action {
		case ActionStart:
			if err := self.StateRef(func(state *GameState) error {
				*state = (*state).GetGame().Start()
				return nil
			}); err != nil {
				return err
			}

		case ActionPlayerChooseWord:
			var action ActionPlayerChooseWordJson
			if err := json.Unmarshal(bytes, &action); err != nil {
				return err
			}

			if err := self.StateRef(func(state *GameState) error {
				game := (*state).(*GamePlayerChooseWord)

				if !CheckSamePlayer(game, game.PlayerIndex2(), playerId) {
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

				if !CheckSamePlayer(game, game.PlayerIndex(), playerId) {
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

				if !CheckSamePlayer(game, game.PlayerIndex(), playerId) {
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

				if !CheckSamePlayer(game, game.PlayerIndex(), playerId) {
					return ErrNotYourTurn
				}

				if newState, ok := game.Vote(action.PlayerIndex); ok {
					*state = newState
				}
				return nil
			}); err != nil {
				return err
			}
		default:
			return ErrUnknownAction
		}

		self.syncGame()
	}
}

func (self *Room) asyncListenStateChange() {
	for {
		select {
		case ok := <-self.stateChangeChan:
			if !ok {
				return
			}
		case <-time.After(time.Second * 30):
			self.stateRefNoSend(func(state *GameState) error {
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

			self.syncGame()
		}
	}
}

func (self *Room) cleanup() {
	close(self.stateChangeChan)
}

func (self *Room) AddPlayer(conn *websocket.Conn, player Player) {
	self.stateNoSend(func(state GameState) error {
		state.GetGame().AddPlayer(player)
		return nil
	})
	self.wsRoom.Add(conn, player.Id)
	self.syncGame()
}

func (self *Room) RemovePlayer(conn *websocket.Conn, playerId string) {
	self.wsRoom.Delete(conn)
	self.stateNoSend(func(state GameState) error {
		if game, ok := state.(*Game); ok {
			game.RemovePlayer(playerId)
		} else {
			state.GetGame().DisconnectPlayer(playerId)
		}
		return nil
	})
	self.syncGame()
}

func (self *Room) syncGame() {
	if err := self.stateNoSend(func(state GameState) error {
		return self.wsRoom.WriteEach(func(writer io.Writer, data any) error {
			return PartialGameState(state, data.(string)).Render(context.Background(), writer)
		})
	}); err != nil {
		log.Println("SyncGame", "err", err)
	}
}

func (self *Room) stateNoSend(mutate func(state GameState) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	return mutate(self.unsafeState)
}

func (self *Room) stateRefNoSend(mutate func(state *GameState) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	return mutate(&self.unsafeState)
}

func (self *Room) StateRef(mutate func(state *GameState) error) error {
	defer func() {
		self.stateChangeChan <- true
	}()

	return self.stateRefNoSend(mutate)
}

func (self *Room) State(mutate func(state GameState) error) error {
	defer func() {
		self.stateChangeChan <- true
	}()

	return self.stateNoSend(mutate)
}
