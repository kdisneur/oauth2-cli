package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateRandomString generates a random string of x number of characters.
func GenerateRandomString(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("can't generate random string: %v", err)
	}

	return hex.EncodeToString(bytes), nil
}
