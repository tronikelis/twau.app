package game_state

import "sync"

type gameWithMutex struct {
	game *Game
	mu   *sync.Mutex
}

// Methods on this struct are concurrency safe
type States struct {
	byId map[string]gameWithMutex
	mu   *sync.Mutex
}

func NewStates() States {
	return States{
		byId: map[string]gameWithMutex{},
		mu:   &sync.Mutex{},
	}
}

func (self States) HasGame(gameId string) bool {
	self.mu.Lock()
	defer self.mu.Unlock()

	_, ok := self.byId[gameId]
	return ok
}

func (self States) Game(gameId string, mutation func(game *Game)) {
	self.mu.Lock()

	game, ok := self.byId[gameId]
	if !ok {
		game.game = NewGame()
		game.mu = &sync.Mutex{}
	}
	self.byId[gameId] = game

	self.mu.Unlock()

	game.mu.Lock()
	mutation(game.game)
	game.mu.Unlock()
}

func (self States) DeleteGame(gameId string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	delete(self.byId, gameId)
}
