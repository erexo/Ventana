package dto

import "github.com/Erexo/Ventana/core/domain"

type User struct {
	Id       int64       `json:"id" db:"id"`
	Username string      `json:"username" db:"username"`
	Role     domain.Role `json:"role" db:"role"`
}
