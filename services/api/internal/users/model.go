package users

import "time"

type User struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	FullName  string    `json:"full_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
