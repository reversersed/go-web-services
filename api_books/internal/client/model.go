package client

import "go.mongodb.org/mongo-driver/bson/primitive"

type Book struct {
	Id        primitive.ObjectID `json:"id" bson:"_id" validate:"primitiveid"`
	Name      string             `json:"name" bson:"name" validate:"min=4,max=32"`
	Author    string             `json:"author" bson:"author" validate:"min=2,max=24"`
	Pages     int                `json:"pages" bson:"-"`
	FirstPage int                `json:"startpage" bson:"startpage"`
	Year      int                `json:"year" bson:"year"`
	FilePath  string             `json:"file" bson:"filepath"`
}
