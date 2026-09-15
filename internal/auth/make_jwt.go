package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	var t *jwt.Token
	rc := jwt.RegisteredClaims {
		Issuer: "chirpy-access",
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject: userId.String(),
	}
	t = jwt.NewWithClaims(jwt.SigningMethodHS256, rc)

	return t.SignedString(
			[]byte(tokenSecret))
}
