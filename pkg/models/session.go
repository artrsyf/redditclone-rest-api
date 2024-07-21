package models

import "time"

type Session struct {
	ID        int       `json:"id"`
	JWT       string    `json:"token"`
	UserID    int       `json:"userId"`
	ExpiresAt time.Time `json:"expires"`
}
