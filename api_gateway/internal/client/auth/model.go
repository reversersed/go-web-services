package auth

type User struct {
	Id    string   `json:"id"`
	Login string   `json:"login"`
	Roles []string `json:"roles"`
}

type UserAuthQuery struct {
	Login    string `json:"login" validate:"required" example:"admin"`
	Password string `json:"password" validate:"required" example:"admin"`
}
