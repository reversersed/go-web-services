package user

type User struct {
	Id             string   `json:"id"`
	Login          string   `json:"login"`
	Roles          []string `json:"roles"`
	Email          string   `json:"email"`
	EmailConfirmed bool     `json:"emailconfirmed"`
}

type UserAuthQuery struct {
	Login    string `json:"login" validate:"required" example:"admin"`
	Password string `json:"password" validate:"required" example:"admin"`
}
type UserRegisterQuery struct {
	Login    string `json:"login" validate:"required,min=4,max=16" example:"user"`
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8,max=32" example:"User!1password"`
}
