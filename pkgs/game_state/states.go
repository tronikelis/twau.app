package game_state

import "sync"

// Methods on this struct are concurrency safe
type States struct {
	byId map[string]*Game
	mu   *sync.Mutex
}

func NewStates() States {
	return States{
		byId: map[string]*Game{},
		mu:   &sync.Mutex{},
	}
}

// returns game or creates a new one
func (self States) Upsert(gameId string) *Game {
	self.mu.Lock()
	defer self.mu.Unlock()

	game, ok := self.byId[gameId]
	if !ok {
		game = NewGame()
	}
	self.byId[gameId] = game

	return game
}

func (self States) Delete(gameId string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	delete(self.byId, gameId)
}
