package types

import (
	"errors"
	"time"
)

type User struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  Password  `json:"-"`
	Activated bool      `json:"activated"`
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

}

// func
