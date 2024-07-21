package models

type Vote struct {
	Author   User `json:"-" bson:"author"`
	AuthorID int  `json:"user,string" bson:"user"`
	Vote     int  `json:"vote" bson:"vote"`
}
