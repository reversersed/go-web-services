package client

type User struct {
	Id       string `json:"id" bson:"_id"`
	Login    string `json:"login" bson:"login"`
	Password []byte `json:"-" bson:"password"`
}

type AuthUserByLoginAndPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
