package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidRole(t *testing.T) {
	assert.True(t, IsValidRole(RoleUser))
	assert.True(t, IsValidRole(RoleAdmin))
	assert.True(t, IsValidRole(RoleModerator))
	assert.True(t, IsValidRole(RoleCourier))
	assert.False(t, IsValidRole("superadmin"))
	assert.False(t, IsValidRole(""))
}

func TestUser_FullName(t *testing.T) {
	user := User{FirstName: "Ivan", LastName: "Petrov"}
	assert.Equal(t, "Ivan Petrov", user.FullName())

	empty := User{}
	assert.Equal(t, "", empty.FullName())
}
