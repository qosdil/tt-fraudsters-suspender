package user

import "time"

type User struct {
	ID        string
	Username  string
	Name      string
	Email     string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
