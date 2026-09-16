package auth

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)


func TestExpiredJwtValidation(t *testing.T) {
	user_id := uuid.New()
	token_secret := os.Getenv("JWT_SECRET")
	expired_expires_in, err := time.ParseDuration("-1s")

	if err != nil {
		t.Fatalf("Time.ParseDuration failed. Check time string!: %v", err)
	}

	json_web_token, err := MakeJWT(user_id, token_secret, expired_expires_in)

	if err != nil {
		t.Fatalf("MakeJWT failed to return a valid web token. Check parameter inputs.: %v", err)
	}

	token_id, err := ValidateJWT(json_web_token, token_secret)

	if err == nil {
		t.Errorf("Token validated despite being older than expiration. TokenID: %v", token_id)
	}
}

func TestInvalidTokenSecretValidation(t *testing.T) {
	user_id := uuid.New()
	token_secret := os.Getenv("JWT_SECRET")
	expires_in, err := time.ParseDuration("1h")

	if err != nil {
		t.Fatalf("Time.ParseDuration failed. Check time string!: %v", err)
	}

	json_web_token, err := MakeJWT(user_id, token_secret, expires_in)

	if err != nil {
		t.Fatalf("json_web_token generation failed. Check input parameters: %v", err)
	}

	nonsense_token_secret := "N0n5ens3Tok3nScret3i98o0vnqaligorpndas#fsa"

	_, err = ValidateJWT(json_web_token, nonsense_token_secret)

	if err == nil {
		t.Errorf("Validation passed despite nonsense secret value being given. NonsenseTokenSecret: %v", nonsense_token_secret)
	}
}




