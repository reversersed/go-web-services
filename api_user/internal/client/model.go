package client

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Login          string             `json:"login" bson:"login"`
	Password       []byte             `json:"-" bson:"password"`
	Roles          []string           `json:"roles" bson:"roles"`
	Email          string             `json:"email" bson:"email"`
	EmailConfirmed bool               `json:"emailconfirmed" bson:"emailconfirmed"`
	LoginCooldown  uint64             `json:"-" bson:"logincooldown"`
}

type AuthUserByLoginAndPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type RegisterUserQuery struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type DeleteUserQuery struct {
	Password string `json:"password"`
}
type ChangeUserLoginQuery struct {
	Login string `json:"newlogin"`
}
