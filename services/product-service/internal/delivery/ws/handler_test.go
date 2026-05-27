package ws

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Notifications_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := NewHub()
	go hub.Run()

	jwtManager := commonjwt.NewManager("secret", time.Hour)
	handler := NewHandler(hub, jwtManager)

	router := gin.New()
	router.GET("/ws/notifications", handler.Notifications)

	req := httptest.NewRequest(http.MethodGet, "/ws/notifications", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_Notifications_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := NewHub()
	go hub.Run()

	jwtManager := commonjwt.NewManager("secret", time.Hour)
	handler := NewHandler(hub, jwtManager)

	router := gin.New()
	router.GET("/ws/notifications", handler.Notifications)

	req := httptest.NewRequest(http.MethodGet, "/ws/notifications?token=not-a-valid-jwt", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_Notifications_TokenFromHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := NewHub()
	go hub.Run()

	jwtManager := commonjwt.NewManager("secret", time.Hour)
	handler := NewHandler(hub, jwtManager)

	router := gin.New()
	router.GET("/ws/notifications", handler.Notifications)

	req := httptest.NewRequest(http.MethodGet, "/ws/notifications", nil)
	req.Header.Set("Authorization", "Bearer also-invalid")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
