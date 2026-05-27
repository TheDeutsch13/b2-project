package config

import (
	"time"

	commonenv "github.com/TheDeutsch13/b2-common/env"
)

type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	UploadDir       string
}

func New() Config {
	return Config{
		Port:            commonenv.GetString("PORT", "8081"),
		DatabaseURL:     commonenv.GetString("DATABASE_URL", "postgres://b2user:b2password@localhost:5432/b2db?sslmode=disable"),
		JWTSecret:       commonenv.GetString("JWT_SECRET", "gamegear-dev-secret-change-me"),
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		UploadDir:       commonenv.GetString("UPLOAD_DIR", "uploads/avatars"),
	}
}
