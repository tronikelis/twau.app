package game_state

import (
	"crypto/rand"
	"encoding/hex"
)

func RandomHex() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
