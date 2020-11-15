package entity

import (
	"database/sql"
	"github.com/Erexo/Ventana/core/domain"
)

type User struct {
	Id       int64          `db:"id"`
	Username string         `db:"username"`
	Password string         `db:"password"`
	Salt     sql.NullString `db:"salt,omitempty"`
	Role     domain.Role    `db:"role"`
}
