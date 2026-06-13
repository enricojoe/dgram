package model

import "time"

// User is a registered account. The password hash is never serialized to JSON.
type User struct {
	ID           int64     `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	DisplayName  string    `json:"displayName" db:"display_name"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

// UserView is the safe, public projection of a User (no password hash).
type UserView struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
}

// View returns the safe public projection of the user.
func (u User) View() UserView {
	return UserView{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName, CreatedAt: u.CreatedAt}
}
