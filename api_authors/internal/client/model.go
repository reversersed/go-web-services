package client

import "go.mongodb.org/mongo-driver/bson/primitive"

type Author struct {
	Id   primitive.ObjectID `json:"id" bson:"_id" validate:"primitiveid"`
	Name string             `json:"name" bson:"name" validate:"min=4,max=32"`
}
