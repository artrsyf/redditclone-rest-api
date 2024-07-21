package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	Created time.Time          `json:"created"`
	Author  *User              `json:"author"`
	Text    string             `json:"body"`
	ID      primitive.ObjectID `json:"id" bson:"_id"`
}
