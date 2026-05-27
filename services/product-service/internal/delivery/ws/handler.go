package ws

import (
	"net/http"
	"strings"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub        *Hub
	jwtManager *commonjwt.Manager
}

func NewHandler(hub *Hub, jwtManager *commonjwt.Manager) *Handler {
	return &Handler{
		hub:        hub,
		jwtManager: jwtManager,
	}
}

func (h *Handler) Notifications(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		authHeader := ctx.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 {
			token = parts[1]
		}
	}

	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	if _, err := h.jwtManager.Parse(token); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	client := make(chan []byte, 16)
	h.hub.Register(client)

	go func() {
		defer func() {
			h.hub.Unregister(client)
			conn.Close()
		}()

		for message := range client {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
