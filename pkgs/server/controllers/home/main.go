package home

import (
	"word-amongus-game/pkgs/server/req"

	"github.com/tronikelis/maruchi"
)

func Register(s *maruchi.Server) {
	s.GET("/{$}", req.WithReqContext(getIndex))
}
