package game_state

import "math/rand/v2"

type words []string

func (self words) RandomN(n int) []string {
	words := make([]string, n)

	for i := range n {
		words[i] = self[rand.IntN(len(allWords))]
	}

	return words
}

var allWords words = words{
	"amongus",
}
