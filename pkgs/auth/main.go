package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// signs "str" bytes and returns a hex signed string
func SignStringHex(str string, key []byte) (string, error) {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write([]byte(str)); err != nil {
		return "", err
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}
