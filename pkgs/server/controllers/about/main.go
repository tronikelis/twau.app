package about

import (
	"github.com/tronikelis/maruchi"
	"twau.app/pkgs/server/req"
)

func Register(s *maruchi.Server) {
	s.GET("/about", req.WithReqContext(getIndex))
}
