package entity

import "database/sql"

type Role int

const (
	RoleNone  Role = 0
	RoleGuest Role = 1
	RoleUser  Role = 2
	RoleAdmin Role = 3
)

type User struct {
	Id       int64          `db:"id"`
	Username string         `db:"username"`
	Password string         `db:"password"`
	Salt     sql.NullString `db:"salt,omitempty"`
	Role     Role           `db:"role"`
}
