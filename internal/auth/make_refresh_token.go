package auth

import (
	"crypto/rand"
	"encoding/hex"
)


func MakeRefreshToken() (refreshToken string, err error) {
	refreshBytes := make([]byte, 32) 
	_, err = rand.Read(refreshBytes)

	if err != nil {
		return "", err
	}
	// hex.EncodeToString transforms the byte to a string
	refreshToken = hex.EncodeToString(refreshBytes)

	return 
}
