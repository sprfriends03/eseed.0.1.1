package util

import (
	"crypto/hmac"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/nhnghia272/gopkg"
	"golang.org/x/crypto/bcrypt"
)

func RandomPassword() string {
	bytes := []byte(gopkg.RandomString(15))
	bytes = append(bytes, []byte(gopkg.RandomNumber(5))...)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(bytes), func(i, j int) { bytes[i], bytes[j] = bytes[j], bytes[i] })
	return string(bytes)
}

func RandomClientId() string {
	return fmt.Sprintf("rci_%v", uuid.NewString())
}

func RandomClientSecret() string {
	return fmt.Sprintf("rcs_%v", uuid.NewString())
}

func RandomSecureKey() string {
	return fmt.Sprintf("rsk_%v", uuid.NewString())
}

func HashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateSignature(data, secret string) string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifySignature(data, secret, signature string) bool {
	payload := CreateSignature(data, secret)
	return subtle.ConstantTimeCompare([]byte(payload), []byte(signature)) == 1
}
