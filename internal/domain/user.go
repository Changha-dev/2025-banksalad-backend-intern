package domain

import (
	"strings"

	"github.com/pkg/errors"
)

type DuplicateStrategy int

const (
	ByEmail DuplicateStrategy = iota
	ByPhone
	ByBoth
)

func (ds DuplicateStrategy) String() string {
	switch ds {
	case ByEmail:
		return "ByEmail"
	case ByPhone:
		return "ByPhone"
	case ByBoth:
		return "ByBoth"
	default:
		return "Unknown"
	}
}

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

func (u *User) UniqueKeyByStrategy(strategy DuplicateStrategy) string {
	switch strategy {
	case ByEmail:
		return u.Email
	case ByPhone:
		return u.PhoneNumber
	case ByBoth:
		return u.Email + "|" + u.PhoneNumber
	default:
		return u.Email
	}
}
