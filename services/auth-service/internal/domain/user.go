package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

const (
	RoleUser      = "user"
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleCourier   = "courier"

	GenderMale   = "male"
	GenderFemale = "female"
)

func IsValidRole(role string) bool {
	switch role {
	case RoleUser, RoleAdmin, RoleModerator, RoleCourier:
		return true
	default:
		return false
	}
}

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
	FirstName    string
	LastName     string
	Nickname     string
	BirthDate    *time.Time
	Gender       string
	Phone        string
	AvatarURL    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u User) FullName() string {
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type UserProfileInput struct {
	FirstName string
	LastName  string
	Nickname  string
	BirthDate *time.Time
	Gender    string
	Phone string
}
