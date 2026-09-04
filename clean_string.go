package main

import "strings"

func CleanString(s string) string {
	profane_words := []string{"kerfuffle", "sharbert", "fornax"}
	input_list := strings.Split(s, " ")

	for i, word := range input_list {
		for _, profane_word := range profane_words {

			if strings.ToLower(word) == profane_word {
				input_list[i] = "****"
			}
		}
	} 

	return strings.Join(input_list, " ")
}
