package util

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token types carried in the custom "typ" claim.
const (
	TokenAccess  = "access"
	TokenRefresh = "refresh"
)

// Token lifetimes.
const (
	accessTTL  = 24 * time.Hour
	refreshTTL = 30 * 24 * time.Hour
)

// ErrInvalidToken is returned when a token is malformed, expired, signed with
// the wrong key, or of an unexpected type.
var ErrInvalidToken = errors.New("invalid token")

// GenerateTokens issues a short-lived access token and a long-lived refresh
// token for the given user, both signed with secret.
func GenerateTokens(userID int64, secret string) (accessToken, refreshToken string, err error) {
	accessToken, err = signToken(userID, TokenAccess, accessTTL, secret)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = signToken(userID, TokenRefresh, refreshTTL, secret)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func signToken(userID int64, typ string, ttl time.Duration, secret string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(userID, 10),
		"typ": typ,
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken validates token against secret, ensures it is of expectedTyp, and
// returns the embedded user id.
func ParseToken(tokenString, secret, expectedTyp string) (int64, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	typ, _ := claims["typ"].(string)
	if typ != expectedTyp {
		return 0, ErrInvalidToken
	}

	sub, _ := claims["sub"].(string)
	userID, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, ErrInvalidToken
	}
	return userID, nil
}
