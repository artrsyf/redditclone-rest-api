package models

type User struct {
	ID       int    `json:"id,string" bson:"id"`
	Login    string `json:"username" bson:"username"`
	Password string `json:"-" bson:"-"`
}
