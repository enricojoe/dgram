// Package repository provides data-access types backed by Postgres via sqlx.
package repository

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/enricojoe/dgram/backend/internal/model"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// UserRepository persists and retrieves users.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository constructs a UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user and populates its generated id and timestamps.
func (r *UserRepository) Create(email, passwordHash string) (model.User, error) {
	var u model.User
	const q = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at`
	err := r.db.Get(&u, q, email, passwordHash)
	return u, err
}

// GetByEmail looks up a user by email. Returns ErrNotFound if none exists.
func (r *UserRepository) GetByEmail(email string) (model.User, error) {
	var u model.User
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`
	if err := r.db.Get(&u, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}
	return u, nil
}

// GetByID looks up a user by id. Returns ErrNotFound if none exists.
func (r *UserRepository) GetByID(id int64) (model.User, error) {
	var u model.User
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`
	if err := r.db.Get(&u, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}
	return u, nil
}
