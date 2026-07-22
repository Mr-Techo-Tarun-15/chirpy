package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() string {
	emptyByteList := make([]byte, 32)
	_, err := rand.Read(emptyByteList)
	if err != nil {
		return err.Error()
	}
	encodedString := hex.EncodeToString(emptyByteList)
	return encodedString
}
