package main

import (
	"fmt"
	"net/http"
)

func JsonError(w http.ResponseWriter, err error) {
	error_response := `{"error": "Something went wrong"}`
	fmt.Printf("Error decoding parameters: %s", err)
	w.WriteHeader(500)
	w.Write([]byte(error_response))
}
