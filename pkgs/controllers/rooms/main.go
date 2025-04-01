package rooms

import "github.com/tronikelis/maruchi"

func Register(s *maruchi.Server) {
	s.POST("/rooms", postIndex)
	s.GET("/rooms/{id}", getRoomId)
}
