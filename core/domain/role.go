package domain

type Role int

const (
	RoleNone  Role = 0
	RoleGuest Role = 1
	RoleUser  Role = 2
	RoleAdmin Role = 3
)
