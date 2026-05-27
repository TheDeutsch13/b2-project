package config

import commonenv "github.com/TheDeutsch13/b2-common/env"

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	UploadDir      string
	CdekClientID   string
	CdekClientSecret string
}

func New() Config {
	return Config{
		Port:             commonenv.GetString("PORT", "8082"),
		DatabaseURL:      commonenv.GetString("DATABASE_URL", "postgres://b2user:b2password@localhost:5432/b2db?sslmode=disable"),
		JWTSecret:        commonenv.GetString("JWT_SECRET", "gamegear-dev-secret-change-me"),
		UploadDir:        commonenv.GetString("UPLOAD_DIR", "uploads"),
		CdekClientID:     commonenv.GetString("CDEK_CLIENT_ID", ""),
		CdekClientSecret: commonenv.GetString("CDEK_CLIENT_SECRET", ""),
	}
}
