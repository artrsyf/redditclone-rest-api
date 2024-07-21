package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	Score            int                `json:"score" bson:"score"`
	Views            int                `json:"views" bson:"views"`
	Type             string             `json:"type" bson:"type"`
	Title            string             `json:"title" bson:"title"`
	Author           User               `json:"author" bson:"author"`
	Category         string             `json:"category" bson:"category"`
	Text             string             `json:"text,omitempty" bson:"text,omitempty"`
	URL              string             `json:"url,omitempty" bson:"url,omitempty"`
	Votes            []*Vote            `json:"votes" bson:"votes"`
	Comments         []*Comment         `json:"comments" bson:"comments"`
	Created          time.Time          `json:"created" bson:"created"`
	UpvotePercentage int                `json:"upvotePercentage" bson:"upvotePercentage"`
	ID               primitive.ObjectID `json:"id" bson:"_id"`
}
