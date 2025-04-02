package game_state

import (
	"sync"

	"word-amongus-game/pkgs/ws"
)

type Room struct {
	State  GameState
	WsRoom *ws.Room
}

// concurrency safe
type Rooms struct {
	statesById map[string]Room
	mu         *sync.Mutex
}

func NewRooms() Rooms {
	return Rooms{
		statesById: map[string]Room{},
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
// WARNING: lock/unlock mutex before modifying the underlying state
func (self Rooms) Room(roomId string) (Room, bool) {
	self.mu.Lock()
	defer self.mu.Unlock()

	game, ok := self.statesById[roomId]
	if !ok {
		return Room{}, false
	}

	return game, true
}

// creates a room, or returns (zero, false) if room exists
func (self Rooms) CreateRoom(roomId string) (Room, bool) {
	if self.HasRoom(roomId) {
		return Room{}, false
	}

	self.mu.Lock()
	defer self.mu.Unlock()

	room := Room{
		State:  NewGame(),
		WsRoom: ws.NewRoom(),
	}

	self.statesById[roomId] = room

	return room, true
}

func (self Rooms) DeleteRoom(gameId string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	delete(self.statesById, gameId)
}
