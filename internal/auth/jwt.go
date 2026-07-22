package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	newJWTToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy-access",
			IssuedAt:  (&jwt.NumericDate{Time: time.Now().UTC()}),
			ExpiresAt: (&jwt.NumericDate{Time: time.Now().Add(expiresIn)}),
			Subject:   userID.String(),
		},
	)

	signedJWT, err := newJWTToken.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedJWT, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	type customClaim struct {
		Foo string
		jwt.RegisteredClaims
	}

	jwtToken, err := jwt.ParseWithClaims(tokenString, &customClaim{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := jwtToken.Claims.(*customClaim)
	if !ok || !jwtToken.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization hader does not exist")
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("authorization header is incomplete")
	}
	return parts[1], nil
}
