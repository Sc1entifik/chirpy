package main

import (
	"time"
	"github.com/google/uuid"
)


type chirp struct {
	ID uuid.UUID `json:"id"`
	Body string `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID uuid.UUID `json:"user_id"`
}
