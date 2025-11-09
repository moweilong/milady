// Package core provides core interfaces and types for jwt
package core

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrRefreshTokenNotFound indicates the refresh token was not found in store
	ErrRefreshTokenNotFound = errors.New("refresh token not found")

	// ErrRefreshTokenExpired indicates the refresh token has expired
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)

// TokenStore defines the interface for storing and retrieving refresh tokens
type TokenStore interface {
	// Set stores a refresh token with associated user data and expiration
	Set(ctx context.Context, token string, userData any, expiry time.Time) error

	// Get retrieves user data associated with a refresh token
	// Returns ErrRefreshTokenNotFound if token does't exist or is expired
	Get(ctx context.Context, token string) (any, error)

	// Delete removes a refresh token from storage
	// Returns an error if the operation fails, but should not error if token doesn't exist
	Delete(ctx context.Context, token string) error

	// Cleanup removes expired tokens (optional, for cleanup routines)
	// Returns the number of tokens
	Cleanup(ctx context.Context) error
}
