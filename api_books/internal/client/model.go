package client

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Author struct {
	Id   primitive.ObjectID `json:"id" validate:"required,primitiveid"`
	Name string             `json:"name"`
}
type Genre struct {
	Id   primitive.ObjectID `json:"id" validate:"required,primitiveid"`
	Name string
}
type Book struct {
	Id        primitive.ObjectID   `json:"id" bson:"_id,omitempty" validate:"primitiveid"`
	Name      string               `json:"name" bson:"name" validate:"min=4,max=32"`
	AuthorId  primitive.ObjectID   `json:"-" bson:"author"`
	Author    *Author              `json:"author,omitempty" bson:"-"`
	GenresId  []primitive.ObjectID `json:"-" bson:"genres"`
	Genres    []*Genre              `json:"genres,omitempty" bson:"-"`
	Pages     int                  `json:"pages" bson:"-"`
	Year      int                  `json:"year" bson:"year"`
	FilePath  string               `json:"file" bson:"filepath"`
	CoverPath string               `json:"cover" bson:"coverpath"`
}

type InsertBookQuery struct {
	Name      string               `validate:"required,min=4,max=32"`
	AuthorId  primitive.ObjectID   `validate:"required,primitiveid"`
	GenresId  []primitive.ObjectID `validate:"required"`
	Year      int                  `validate:"gte=1400,lte=2100"`
	Pages     int                  `validate:"required,lte=5000"`
	FilePath  string
	CoverPath string
}
