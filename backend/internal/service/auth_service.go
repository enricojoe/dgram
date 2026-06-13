package service

import (
	"errors"
	"strings"

	"github.com/enricojoe/dgram/backend/internal/model"
	"github.com/enricojoe/dgram/backend/internal/repository"
	"github.com/enricojoe/dgram/backend/internal/util"
)

// Sentinel errors returned by AuthService so controllers can map them to the
// correct HTTP status without depending on lower layers.
var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// AuthService handles registration, login and token refresh.
type AuthService struct {
	users     *repository.UserRepository
	jwtSecret string
}

// NewAuthService constructs an AuthService.
func NewAuthService(users *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret}
}

// Tokens bundles an issued access/refresh token pair.
type Tokens struct {
	AccessToken  string
	RefreshToken string
}

// Register creates a new account, returning the user and a fresh token pair.
// Returns ErrEmailTaken if the email is already in use.
func (s *AuthService) Register(email, password string) (model.User, Tokens, error) {
	if _, err := s.users.GetByEmail(email); err == nil {
		return model.User{}, Tokens{}, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return model.User{}, Tokens{}, err
	}

	hash, err := util.HashPassword(password)
	if err != nil {
		return model.User{}, Tokens{}, err
	}

	user, err := s.users.Create(email, hash)
	if err != nil {
		return model.User{}, Tokens{}, err
	}

	tokens, err := s.issue(user.ID)
	if err != nil {
		return model.User{}, Tokens{}, err
	}
	return user, tokens, nil
}

// Login verifies credentials and returns the user with a fresh token pair.
// Returns ErrInvalidCredentials on any mismatch (without leaking which part).
func (s *AuthService) Login(email, password string) (model.User, Tokens, error) {
	user, err := s.users.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.User{}, Tokens{}, ErrInvalidCredentials
		}
		return model.User{}, Tokens{}, err
	}

	if !util.CheckPassword(user.PasswordHash, password) {
		return model.User{}, Tokens{}, ErrInvalidCredentials
	}

	tokens, err := s.issue(user.ID)
	if err != nil {
		return model.User{}, Tokens{}, err
	}
	return user, tokens, nil
}

// Refresh validates a refresh token and issues a new token pair.
func (s *AuthService) Refresh(refreshToken string) (Tokens, error) {
	userID, err := util.ParseToken(refreshToken, s.jwtSecret, util.TokenRefresh)
	if err != nil {
		return Tokens{}, ErrInvalidCredentials
	}
	// Ensure the user still exists.
	if _, err := s.users.GetByID(userID); err != nil {
		return Tokens{}, ErrInvalidCredentials
	}
	return s.issue(userID)
}

func (s *AuthService) issue(userID int64) (Tokens, error) {
	access, refresh, err := util.GenerateTokens(userID, s.jwtSecret)
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{AccessToken: access, RefreshToken: refresh}, nil
}

// GetUser returns the user with the given id.
func (s *AuthService) GetUser(id int64) (model.User, error) {
	return s.users.GetByID(id)
}

// UpdateDisplayName updates the user's display name and returns the updated user.
func (s *AuthService) UpdateDisplayName(id int64, displayName string) (model.User, error) {
	return s.users.UpdateDisplayName(id, strings.TrimSpace(displayName))
}

// UpdatePassword changes the user's password after verifying the current one.
// Returns ErrInvalidCredentials if the current password does not match.
func (s *AuthService) UpdatePassword(id int64, oldPassword, newPassword string) error {
	user, err := s.users.GetByID(id)
	if err != nil {
		return err
	}

	if !util.CheckPassword(user.PasswordHash, oldPassword) {
		return ErrInvalidCredentials
	}

	hash, err := util.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.users.UpdatePasswordHash(id, hash)
}
