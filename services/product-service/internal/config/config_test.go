package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_DefaultValues(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")

	cfg := New()

	assert.Equal(t, "8082", cfg.Port)
	assert.Equal(t, "postgres://b2user:b2password@localhost:5432/b2db?sslmode=disable", cfg.DatabaseURL)
}

func TestNew_FromEnv(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb?sslmode=disable")

	cfg := New()

	assert.Equal(t, "9000", cfg.Port)
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb?sslmode=disable", cfg.DatabaseURL)
}
