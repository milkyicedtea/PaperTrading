package user

import (
	"github.com/google/uuid"
	"time"
)

// User represents a user in the system.
type User struct {
	ID           uuid.UUID `json:"id" db:"id"` // use string for flexibility (e.g. UUID)
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// NewUser contains information needed to create a new user.
// might be used this in the service layer.
type NewUser struct {
	Email    string
	Password string // Plain text password that will be hashed
}
