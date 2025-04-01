package controllers

import (
	"word-amongus-game/pkgs/controllers/home"

	"github.com/tronikelis/maruchi"
)

func Register(s *maruchi.Server) {
	home.Register(s)
}
