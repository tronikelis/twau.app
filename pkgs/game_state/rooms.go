package game_state

import (
	"sync"

	"word-amongus-game/pkgs/ws"
)

type Room struct {
	WsRoom      *ws.Room
	UnsafeState GameState
	mu          *sync.Mutex
}

func NewRoom() *Room {
	return &Room{
		WsRoom:      ws.NewRoom(),
		UnsafeState: NewGame(),
		mu:          &sync.Mutex{},
	}
}

func (self *Room) StateRef(mutate func(state *GameState) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	return mutate(&self.UnsafeState)
}

func (self *Room) State(mutate func(state GameState) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	return mutate(self.UnsafeState)
}

// concurrency safe
type Rooms struct {
	statesById map[string]*Room
	mu         *sync.Mutex
}

func NewRooms() Rooms {
	return Rooms{
		statesById: map[string]*Room{},
		mu:         &sync.Mutex{},
	}
}

func (self Rooms) HasRoom(roomId string) bool {
	self.mu.Lock()
	defer self.mu.Unlock()

	_, ok := self.statesById[roomId]
	return ok
}

// gets room or returns (zero, false)
func (self Rooms) Room(roomId string) (*Room, bool) {
	self.mu.Lock()
	defer self.mu.Unlock()

	game, ok := self.statesById[roomId]
	if !ok {
		return nil, false
	}

	return game, true
}

// creates a room, or returns (zero, false) if room exists
func (self Rooms) CreateRoom(roomId string) (*Room, bool) {
	if self.HasRoom(roomId) {
		return nil, false
	}

	self.mu.Lock()
	defer self.mu.Unlock()

	room := NewRoom()
	self.statesById[roomId] = room

	return room, true
}

func (self Rooms) DeleteRoom(roomId string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	delete(self.statesById, roomId)
}
