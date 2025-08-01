package game_state

import (
	"log"
	"sync"
	"time"

	"twau.app/pkgs/random"
)

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
		room.cancelDeleteChan = nil
	}

	return room.room, true
}

func (self Rooms) CreateRoom(password string) (*Room, string) {
	var roomId string
	for {
		roomId = random.RandomB64(6)
		if !self.HasRoom(roomId) {
			break
		}
	}

	self.mu.Lock()
	defer self.mu.Unlock()

	room := NewRoom(password)
	self.statesById[roomId] = &roomWithDeleteChan{
		room: room,
	}

	return room, roomId
}

func (self Rooms) QueueDelete(roomId string) {
	log.Println("queueing deletion of", roomId)

	self.mu.Lock()
	defer self.mu.Unlock()

	room, ok := self.statesById[roomId]
	if !ok {
		return
	}

	if room.cancelDeleteChan != nil {
		close(room.cancelDeleteChan)
		room.cancelDeleteChan = nil
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

	room, ok := self.statesById[roomId]
	if !ok {
		return
	}

	room.room.cleanup()
	delete(self.statesById, roomId)
}
