package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// signs "str" bytes and returns a hex signed string
func SignStringB64(str string, key []byte) (string, error) {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write([]byte(str)); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(mac.Sum(nil)), nil
}
