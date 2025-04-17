package game_state

import "math/rand/v2"

type words []string

func (self words) randomN(n int) []string {
	words := make([]string, n)

	for i := range n {
		words[i] = self[rand.IntN(len(allWords))]
	}

	return words
}

var allWords words = words{
	"apple",
	"banana",
	"car",
	"dog",
	"elephant",
	"forest",
	"guitar",
	"house",
	"island",
	"jungle",
	"kite",
	"lamp",
	"mountain",
	"notebook",
	"ocean",
	"piano",
	"queen",
	"river",
	"sun",
	"tree",
	"umbrella",
	"violin",
	"whale",
	"xylophone",
	"yacht",
	"zebra",
}
