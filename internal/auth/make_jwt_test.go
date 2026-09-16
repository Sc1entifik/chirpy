package auth

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	user_id := uuid.New()
	token_secret := os.Getenv("JWT_SECRET")
	expires_in, err := time.ParseDuration("1h")

	if err != nil {
		t.Fatalf("failed to parse duration: %v", err)
	}

	json_web_token, err := MakeJWT(user_id, token_secret, expires_in)

	if err != nil {
		t.Errorf("Test failed to make a proper JWT: %v", err)
	}

	token_id, err := ValidateJWT(json_web_token, token_secret)

	if err != nil {
		t.Errorf("Test failed to validate the JWT: %v", err)
	}

	if token_id != user_id {
		t.Errorf("Test failed wrong ID value.\nuser_id: %v\ntoken_id: %v",user_id, token_id)
	}
}
