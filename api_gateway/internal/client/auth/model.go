package auth

type User struct {
	Login string
}

type UserAuthQuery struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}
