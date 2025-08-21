package entity

type UserDTO struct {
	ID           int
	Login        string
	PasswordHash string
	Salt         string
}

type AuthData struct {
	Salt     string
	JWTToken string
}
