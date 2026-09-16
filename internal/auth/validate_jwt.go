package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func ValidateJWT(tokenString string, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil 
	})

	if err != nil {
		return uuid.Nil , err
	}
	claims_id, err := uuid.FromBytes([]byte(claims.Subject))

	if err != nil {
		return uuid.Nil, err
	}

	return claims_id, err
}
