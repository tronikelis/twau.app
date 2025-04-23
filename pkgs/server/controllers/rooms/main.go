package rooms

import (
	"twau.app/pkgs/server/req"

	"github.com/tronikelis/maruchi"
)

func Register(s *maruchi.Server) {
	s.POST("/rooms/{id}", req.WithReqContext(postId))
	s.GET("/rooms/{id}", req.WithReqContext(getId))
	s.GET("/rooms/{id}/ws", req.WithReqContext(wsId))
	s.POST("/rooms", req.WithReqContext(postIndex))
}
