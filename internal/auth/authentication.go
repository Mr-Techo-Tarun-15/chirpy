package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(
		password,
		(&argon2id.Params{
			Memory:      64 * 1024,
			Iterations:  1,
			Parallelism: 2,
			SaltLength:  16,
			KeyLength:   32,
		}),
	)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	pswrdMatches, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return pswrdMatches, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization hader does not exist")
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "ApiKey") {
		return "", fmt.Errorf("authorization header is incomplete")
	}
	return parts[1], nil
}
