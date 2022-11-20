package utils

import (
	"crypto/aes"
	base64 "encoding/base64"
	"fmt"
)

func EncryptAES(key []byte, plaintext string) (string, error) {

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	out := make([]byte, len(plaintext))

	c.Encrypt(out, []byte(plaintext))

	// return hex.EncodeToString(out), nil
	return base64.RawURLEncoding.EncodeToString(out), nil
}

func DecryptAES(key []byte, ct string) (string, error) {
	// ciphertext, errDecode := hex.DecodeString(ct)
	ciphertext, errDecode := base64.StdEncoding.DecodeString(ct)
	if errDecode != nil {
		return "", errDecode
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	pt := make([]byte, len(ciphertext))
	c.Decrypt(pt, ciphertext)

	s := string(pt[:])
	fmt.Println("DECRYPTED:", s)
	return s, nil
}
