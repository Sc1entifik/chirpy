package auth

import (
	"net/http"
	"strings"
	"testing"
)


func TestGetBearerToken(t *testing.T) {
	header := make(http.Header)

	bearer_token := "Bearer eiovn3#Q#vvfol93ng30qfafdas"

	header.Add("Authorization", bearer_token)

	parsed_bearer_token, err := GetBearerToken(header)

	if err != nil {
		t.Errorf("Initial parsing of bearer token failed: %v", err)
	}

	trimmed_bearer_token := strings.TrimPrefix(bearer_token, "Bearer ")

	if trimmed_bearer_token != parsed_bearer_token {
		t.Errorf("Parsed bearer token does not match test:\ntrimmed_bearer_token: %v\nparsed_bearer_token: %v", trimmed_bearer_token, parsed_bearer_token)
	}
}
