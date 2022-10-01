package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// ============== SHA256 to string ==============
func GetSHA256Hash(text string) string {
	data := []byte(text)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
