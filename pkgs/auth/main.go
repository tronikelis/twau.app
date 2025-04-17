package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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

// signs "str" bytes and returns a hex signed string
func SignStringHex(str string, key []byte) (string, error) {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write([]byte(str)); err != nil {
		return "", err
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}
