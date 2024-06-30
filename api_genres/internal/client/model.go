package client

import "go.mongodb.org/mongo-driver/bson/primitive"

type Genre struct {
	Id   primitive.ObjectID `json:"id" bson:"_id,omitempty" validate:"primitiveid"`
	Name string             `json:"name" bson:"name" validate:"min=4,max=32"`
}

type AddGenreQuery struct {
	Name string `json:"name" validate:"min=4,max=32"`
}
