package http

import (
	"time"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
)

func authBearerToken(userID int64, email, role string) string {
	jwtManager := commonjwt.NewManager("test-secret", time.Hour)
	token, _ := jwtManager.Generate(userID, email, role)
	return "Bearer " + token
}
