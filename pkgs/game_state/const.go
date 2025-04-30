package game_state

import "math/rand/v2"

func randomN[T any](all []T, n int) []T {
	if n >= len(all) {
		return all
	}

	buf := make([]T, n)

	for i := range n {
		buf[i] = all[rand.IntN(len(all))]
	}

	return buf
}

type Category struct {
	Id    int
	Name  string
	Words []string
}

var buildingWords = []string{
	"school",
	"hospital",
	"church",
	"library",
	"courthouse",
	"police station",
	"fire station",
	"post-office",
	"museum",
	"airport",
	"train-station",
	"city hall",
	"bank",
	"supermarket",
	"warehouse",
	"factory",
	"theater",
	"stadium",
	"hotel",
	"apartment building",
}

var animalWords = []string{
	"penguin (linuxer) 🐧❤️",
	"dog",
	"cat",
	"lion",
	"tiger",
	"elephant",
	"giraffe",
	"zebra",
	"monkey",
	"bear",
	"wolf",
	"fox",
	"rabbit",
	"deer",
	"kangaroo",
	"panda",
	"dolphin",
	"whale",
	"eagle",
	"owl",
	"snake",
	"turtle",
}

var toolWords = []string{
	"hammer",
	"screwdriver",
	"wrench",
	"pliers",
	"saw",
	"drill",
	"level",
	"tape-measure",
	"axe",
	"hoe",
}

var apparelWords = []string{
	"shirt",
	"pants",
	"jacket",
	"coat",
	"dress",
	"skirt",
	"blouse",
	"sweater",
	"jeans",
	"shorts",
	"hat",
	"scarf",
	"gloves",
	"socks",
	"shoes",
	"boots",
	"belt",
	"tie",
	"hoodie",
	"pajamas",
}

var allCategories = []Category{
	{
		Id:    1,
		Name:  "Buildings",
		Words: buildingWords,
	},
	{
		Id:    2,
		Name:  "Animals",
		Words: animalWords,
	},
	{
		Id:    3,
		Name:  "Tools",
		Words: toolWords,
	},
	{
		Id:    4,
		Name:  "Apparel",
		Words: apparelWords,
	},
}
