package req

import (
	"context"
	// "sync"

	"word-amongus-game/pkgs/game_state"

	// "github.com/gorilla/websocket"
	"github.com/tronikelis/maruchi"
)

const (
	ContextKey int = iota
)

// // methods are concurency safe
// type WsByPlayerId struct {
// 	mu           *sync.Mutex
// 	wsByPlayerId map[string]*websocket.Conn
// }
//
// // map[key]
// func (self WsByPlayerId) Load(playerId string) (*websocket.Conn, bool) {
// 	self.mu.Lock()
// 	defer self.mu.Unlock()
//
// 	ws, ok := self.wsByPlayerId[playerId]
// 	return ws, ok
// }
//
// func (self WsByPlayerId) Insert(playerId string, ws *websocket.Conn) {
// 	self.mu.Lock()
// 	defer self.mu.Unlock()
// 	self.wsByPlayerId[playerId] = ws
// }
//
// func (self WsByPlayerId) Delete(playerId string) {
// 	self.mu.Lock()
// 	defer self.mu.Unlock()
// 	delete(self.wsByPlayerId, playerId)
// }
//
// func NewWsByPlayerId() WsByPlayerId {
// 	return WsByPlayerId{
// 		mu:           &sync.Mutex{},
// 		wsByPlayerId: map[string]*websocket.Conn{},
// 	}
// }

type ReqContext struct {
	Rooms game_state.Rooms
	// WsByPlayerId WsByPlayerId
}

func NewReqContext() ReqContext {
	return ReqContext{
		Rooms: game_state.NewRooms(),
		// WsByPlayerId: NewWsByPlayerId(),
	}
}

func InitContext(request *maruchi.ReqContextBase, reqContext ReqContext) {
	newContext := context.WithValue(
		request.Context(),
		ContextKey,
		reqContext,
	)

	request.R = request.R.WithContext(newContext)
}

func GetReqContext(ctx maruchi.ReqContext) ReqContext {
	return ctx.Context().Value(ContextKey).(ReqContext)
}
