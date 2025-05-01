package random

import (
	"crypto/rand"
	"encoding/hex"
	rand2 "math/rand/v2"
)

const (
	LengthPlayerId = 16
	LengthRoomId   = 16
)

func RandomHex(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

type NormalizedRandom struct {
	counts []int
}

func NewNormalizedRandom(n int) NormalizedRandom {
	counts := make([]int, n)
	for i := range counts {
		// make sure to initialize counts to non 0, as they will be divided by
		counts[i] = rand2.IntN(5) + 1
	}

	return NormalizedRandom{
		counts: counts,
	}
}

func (self NormalizedRandom) increment(i int) {
	prev := self.counts[i]
	self.counts[i] = prev + 1
}

// [0.05, 0.95]
func (self NormalizedRandom) normalized() []float64 {
	highest := -1
	for _, v := range self.counts {
		if v > highest {
			highest = v
		}
	}

	normalized := make([]float64, len(self.counts))
	for i, v := range self.counts {
		normalized[i] = 1 - (float64(v) / float64(highest))

		switch {
		case normalized[i] < 0.05:
			normalized[i] = 0.05
		case normalized[i] > 0.95:
			normalized[i] = 0.95
		}
	}

	return normalized
}

// returns [0, n)
func (self NormalizedRandom) Int() int {
	normalized := self.normalized()

	highestI := 0
	highestV := float64(-1)

	randFloat := rand2.Float64()
	for i, v := range normalized {
		if highestV < v {
			highestV = v
			highestI = i
		}

		if randFloat <= v {
			self.increment(i)
			return i
		}
	}

	self.increment(highestI)
	return highestI
}
