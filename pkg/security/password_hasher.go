package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordMismatch = errors.New("password mismatch")
)

// PasswordHasher defines the contract for password hashing operations.
// mockery --name=PasswordHasher --dir=pkg/security --output=pkg/security/mocks --filename=password_hasher.go
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type bcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher creates a new bcrypt-based password hasher.
func NewBcryptPasswordHasher() PasswordHasher {
	return &bcryptPasswordHasher{
		cost: bcrypt.DefaultCost,
	}
}

// Hash generates a bcrypt hash from the given password string.
func (h *bcryptPasswordHasher) Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Compare compares a bcrypt hashed password with its possible plaintext equivalent.
// Returns ErrPasswordMismatch if the password doesn't match, or other errors for system failures.
func (h *bcryptPasswordHasher) Compare(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == nil {
		return nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrPasswordMismatch
	}
	return err
}
