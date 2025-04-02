package ui

import (
	"github.com/a-h/templ"
)

// string(templ.URL(url))
func StringURL(url string) string {
	return string(templ.URL(url))
}
