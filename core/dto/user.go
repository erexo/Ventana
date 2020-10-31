package dto

import "github.com/Erexo/Ventana/core/entity"

type User struct {
	Id       int64       `json:"id" db:"id"`
	Username string      `json:"username" db:"username"`
	Role     entity.Role `json:"role" db:"role"`
}