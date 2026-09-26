package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetApiKey(headers http.Header) (apiKey string, err error) {
	apiKey = headers.Get("Authorization")
	err = nil

	if apiKey == "" {
		err = errors.New("Authorization Header Not Present")
		return
	}

	apiKey = strings.TrimPrefix(apiKey, "ApiKey ")
	return
}
