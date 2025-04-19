package game_state

import (
	"log"
	"sync"
	"time"

	"word-amongus-game/pkgs/ws"
)

type Room struct {
	wsRoom      *ws.Room
	unsafeState GameState
	mu          *sync.Mutex
}

func NewRoom() *Room {
	return &Room{
		wsRoom:      ws.NewRoom(),
		unsafeState: NewGame(),
		mu:          &sync.Mutex{},
	}
}

func (self *Room) WsRoom() *ws.Room {
	return self.wsRoom
}

func (self *Room) StateRef(mutate func(state *GameState) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	return mutate(&self.unsafeState)
}

func (self *Room) State(mutate func(state GameState) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	return mutate(self.unsafeState)
}

type roomWithDeleteChan struct {
	room             *Room // can be returned with a locked outer mutex
	cancelDeleteChan chan struct{}
}

// concurrency safe
type Rooms struct {
	statesById map[string]*roomWithDeleteChan
	mu         *sync.Mutex
}

func NewRooms() Rooms {
	return Rooms{
		statesById: map[string]*roomWithDeleteChan{},
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

	room, ok := self.statesById[roomId]
	if !ok {
		return nil, false
	}

	if room.cancelDeleteChan != nil {
		close(room.cancelDeleteChan)
	}

	return room.room, true
}

// creates a room, or returns (zero, false) if room exists
func (self Rooms) CreateRoom(roomId string) (*Room, bool) {
	if self.HasRoom(roomId) {
		return nil, false
	}

	self.mu.Lock()
	defer self.mu.Unlock()

	room := NewRoom()
	self.statesById[roomId] = &roomWithDeleteChan{
		room: room,
	}

	return room, true
}

func (self Rooms) QueueDelete(roomId string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	room, ok := self.statesById[roomId]
	if !ok {
		return
	}

	cancelDeleteChan := make(chan struct{})
	room.cancelDeleteChan = cancelDeleteChan

	go func() {
		select {
		case <-time.After(time.Second * 30):
			self.deleteRoom(roomId)
			return
		case <-cancelDeleteChan:
			return
		}
	}()
}

func (self Rooms) deleteRoom(roomId string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	log.Println("deleting room", roomId)

	delete(self.statesById, roomId)
}
