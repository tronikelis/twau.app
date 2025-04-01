package home

import "github.com/tronikelis/maruchi"

func Register(s *maruchi.Server) {
	s.GET("/{$}", getIndex)
}
