package types

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrRecordNotFound = errors.New("record not found")
)

type User struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  Password  `json:"-"`
	Version   int       `json:"-"`
}

type Password struct {
	Plaintext *string
	Hash      []byte
}

func (p Password) ValidatePassword(value string) error {
	if value == "" {
		return errors.New("Password must be Provided")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(value), 12)
	if err != nil {
		return err
	}
	p.Plaintext = &value
	p.Hash = hash
	return nil
}

func (p Password) Matches(text string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(text))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, err
		default:
			return false, err
		}
	}
	return true, nil
}
