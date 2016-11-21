package steamcommunity

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

func generateSessionID() (string, error) {
	b := make([]byte, 12)
	n, err := rand.Read(b)
	if n != 12 || err != nil {
		return "", errors.New("Failed to generate SessionID")
	}
	return hex.EncodeToString(b), nil
}
