package calltypes

import "time"

// User provides structure to hold users
// @Description info about user.
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName,omitempty"`
	LastName  string    `json:"lastName,omitempty"`
	Password  string    `json:"-"`
	Active    int       `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// LoginRequest represents user login request
// @name LoginRequest.
type LoginRequest struct {
	Email    string `example:"user@example.com"    json:"email"`
	Password string `example:"securePassword123"   json:"password"`
}

// RegisterRequest represents user registration request
// @name RegisterRequest.
type RegisterRequest struct {
	Email     string `example:"user@example.com"    json:"email"`
	FirstName string `example:"John"                json:"firstName"`
	LastName  string `example:"Doe"                 json:"lastName"`
	Password  string `example:"securePassword123"   json:"password"`
	Active    int    `example:"1"                   json:"active,omitempty"`
}
