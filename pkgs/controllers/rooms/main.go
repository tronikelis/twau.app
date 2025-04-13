package rooms

import "github.com/tronikelis/maruchi"

func Register(s *maruchi.Server) {
	s.POST("/rooms/{id}", postId)
	s.GET("/rooms/{id}", getId)
	s.GET("/rooms/{id}/ws", wsId)
	s.POST("/rooms", postIndex)
}
