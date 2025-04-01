package controllers

import (
	"word-amongus-game/pkgs/controllers/home"
	"word-amongus-game/pkgs/controllers/rooms"

	"github.com/tronikelis/maruchi"
)

func Register(s *maruchi.Server) {
	home.Register(s)
	rooms.Register(s)
}
