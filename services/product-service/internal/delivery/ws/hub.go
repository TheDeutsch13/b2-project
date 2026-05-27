package ws

import (
	"encoding/json"
	"sync"
)

type Notification struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	OrderID int64  `json:"order_id,omitempty"`
	UserID  int64  `json:"user_id,omitempty"`

	SupportThreadID   int64  `json:"support_thread_id,omitempty"`
	SupportMessageID  int64  `json:"support_message_id,omitempty"`
	SupportBody       string `json:"support_body,omitempty"`
	SupportSenderID   int64  `json:"support_sender_id,omitempty"`
	SupportSenderRole string `json:"support_sender_role,omitempty"`
	SupportSenderName string `json:"support_sender_name,omitempty"`
	TargetUserID      int64  `json:"target_user_id,omitempty"`
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[chan []byte]struct{}
	broadcast  chan []byte
	register   chan chan []byte
	unregister chan chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[chan []byte]struct{}),
		broadcast:  make(chan []byte, 64),
		register:   make(chan chan []byte),
		unregister: make(chan chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client <- message:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Broadcast(notification Notification) {
	payload, err := json.Marshal(notification)
	if err != nil {
		return
	}

	h.broadcast <- payload
}

func (h *Hub) Register(client chan []byte) {
	h.register <- client
}

func (h *Hub) Unregister(client chan []byte) {
	h.unregister <- client
}
