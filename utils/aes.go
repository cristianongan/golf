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
	ptUnPadding := pkcs7Unpadding(pt)
	s := string(ptUnPadding[:])
	fmt.Println("DECRYPTED:", s)
	return s, nil
}

func pkcs7Unpadding(src []byte) []byte {
	length := len(src)
	unPadding := int(src[length-1])
	return src[:(length - unPadding)]
}
