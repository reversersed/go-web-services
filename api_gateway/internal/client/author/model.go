package author

import "go.mongodb.org/mongo-driver/bson/primitive"

type Author struct {
	Id   primitive.ObjectID `json:"id" validate:"required,primitiveid"`
	Name string             `json:"name"`
}
