package controllers

import (
	"word-amongus-game/pkgs/server/controllers/home"
	"word-amongus-game/pkgs/server/controllers/rooms"

	"github.com/tronikelis/maruchi"
)

func Register(s *maruchi.Server) {
	home.Register(s)
	rooms.Register(s)
}
