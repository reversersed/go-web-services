package genre

import "go.mongodb.org/mongo-driver/bson/primitive"

type Genre struct {
	Id   primitive.ObjectID `json:"id" validate:"required,primitiveid"`
	Name string             `json:"name"`
}

type AddGenreQuery struct {
	Name string `json:"name"`
}
