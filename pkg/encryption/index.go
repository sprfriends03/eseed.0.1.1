package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

func Encrypt(data, secret string) string {
	key := make([]byte, 32)
	copy(key, []byte(secret))
	bytes := []byte(data)
	block, _ := aes.NewCipher(key)
	aesGCM, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesGCM.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	byteEncrypted := aesGCM.Seal(nonce, nonce, bytes, nil)
	return fmt.Sprintf("%x", byteEncrypted)
}

func Decrypt(encrypted, secret string) string {
	key := make([]byte, 32)
	copy(key, []byte(secret))
	bytes, _ := hex.DecodeString(encrypted)
	block, _ := aes.NewCipher(key)
	aesGCM, _ := cipher.NewGCM(block)
	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := bytes[:nonceSize], bytes[nonceSize:]
	byteData, _ := aesGCM.Open(nil, nonce, ciphertext, nil)
	return string(byteData)
}
