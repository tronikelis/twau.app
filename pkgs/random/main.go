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

type RandomIntNotSame struct {
	previous int
	margin   int
}

// margin tells how many times to try to get a different int than the previous,
// increasing the value decreases the chance to get the same number
func NewRandomIntNotSame(margin int) RandomIntNotSame {
	return RandomIntNotSame{
		previous: -1,
		margin:   margin,
	}
}

func (self *RandomIntNotSame) IntN(n int) int {
	var randomInt int

	for range self.margin + 1 {
		randomInt = rand2.IntN(n)
		if randomInt != self.previous {
			self.previous = randomInt
			return randomInt
		}
	}

	self.previous = randomInt
	return randomInt
}
