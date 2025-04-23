package controllers

import (
	"twau.app/pkgs/server/controllers/home"
	"twau.app/pkgs/server/controllers/rooms"

	"github.com/tronikelis/maruchi"
)

func Register(s *maruchi.Server) {
	home.Register(s)
	rooms.Register(s)
}
