package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	bearer_token := headers.Get("Authorization")

	if bearer_token == "" {
		return "", errors.New("Authorization Header Not Present")
	}

	return strings.TrimPrefix(bearer_token, "Bearer "), nil
}
