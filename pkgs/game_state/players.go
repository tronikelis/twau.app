package game_state

import (
	"container/list"
	"fmt"
)

type Player struct {
	Id     string
	Name   string
	Online bool
}

func NewPlayer(id string, name string) Player {
	return Player{Id: id, Name: name, Online: true}
}

type Players struct {
	playerPtrs map[string]*list.Element
	players    *list.List
}

func newPlayers() Players {
	return Players{
		players:    list.New(),
		playerPtrs: make(map[string]*list.Element),
	}
}

func (self Players) Add(player Player) {
	playerPtr, ok := self.playerPtrs[player.Id]
	if ok {
		playerPtr.Value = player
		return
	}

	self.playerPtrs[player.Id] = self.players.PushBack(player)
}

func (self Players) Remove(id string) {
	playerPtr, ok := self.playerPtrs[id]
	if !ok {
		return
	}

	self.players.Remove(playerPtr)
	delete(self.playerPtrs, id)
}

func (self Players) Disconnect(id string) {
	playerPtr, ok := self.playerPtrs[id]
	if !ok {
		return
	}

	prev := playerPtr.Value.(Player)
	prev.Online = false
	playerPtr.Value = prev
}

func (self Players) deleteFunc(fn func(v Player) bool) {
	for k, v := range self.playerPtrs {
		if !fn(v.Value.(Player)) {
			continue
		}

		self.players.Remove(v)
		delete(self.playerPtrs, k)
	}
}

func (self Players) ClearOffline() {
	self.deleteFunc(func(v Player) bool {
		return !v.Online
	})
}

func (self Players) Online() int {
	count := 0
	for v := self.players.Front(); v != nil; v = v.Next() {
		if v.Value.(Player).Online {
			count++
		}
	}
	return count
}

func (self Players) Player(id string) (Player, bool) {
	player, ok := self.playerPtrs[id]
	if !ok {
		return Player{}, false
	}

	return player.Value.(Player), true
}

func (self Players) PlayerOrPanic(id string) Player {
	p, ok := self.Player(id)
	if !ok {
		panic("expected player to exist")
	}
	return p
}

func (self Players) Players() []Player {
	players := make([]Player, 0, self.Len())

	for v := self.players.Front(); v != nil; v = v.Next() {
		players = append(players, v.Value.(Player))
	}

	return players
}

func (self Players) Index(i int) Player {
	for j, v := 0, self.players.Front(); v != nil; j, v = j+1, v.Next() {
		if j == i {
			return v.Value.(Player)
		}
	}

	panic(fmt.Sprintf("Index called out of bounds with %d, len is %d", i, self.Len()))
}

func (self Players) Len() int {
	return self.players.Len()
}

func (self Players) NextFrom(id string) Player {
	player, ok := self.playerPtrs[id]
	if !ok {
		return Player{}
	}

	player = player.Next()
	if player == nil {
		player = self.players.Front()
	}

	return player.Value.(Player)
}
