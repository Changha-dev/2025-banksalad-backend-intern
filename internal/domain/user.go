package domain

import (
	"errors"
	"strings"
)

type User struct {
	Email       string
	PhoneNumber string
	CreditUp    bool
}

func NewUser(email, phoneNumber string, creditUp bool) (*User, error) {
	email = strings.TrimSpace(email)
	phoneNumber = strings.TrimSpace(phoneNumber)

	if len(email) == 0 {
		return nil, errors.New("email cannot be empty")
	}

	if len(phoneNumber) == 0 {
		return nil, errors.New("phone number cannot be empty")
	}

	return &User{
		Email:       email,
		PhoneNumber: phoneNumber,
		CreditUp:    creditUp,
	}, nil
}

func (u *User) IsEligibleForNotification() bool {
	return u.CreditUp
}

func (u *User) UniqueKey() string {
	return u.Email
}
