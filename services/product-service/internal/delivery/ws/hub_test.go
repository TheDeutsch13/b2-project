package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := make(chan []byte, 1)
	hub.Register(client)

	time.Sleep(20 * time.Millisecond)

	hub.Broadcast(Notification{
		Type:    "test",
		Message: "hello",
	})

	select {
	case payload := <-client:
		var notification Notification
		assert.NoError(t, json.Unmarshal(payload, &notification))
		assert.Equal(t, "hello", notification.Message)
	case <-time.After(time.Second):
		t.Fatal("expected notification")
	}
}
