package home

import (
	"twau.app/pkgs/server/req"

	"github.com/tronikelis/maruchi"
)

func Register(s *maruchi.Server) {
	s.GET("/{$}", req.WithReqContext(getIndex))
}
