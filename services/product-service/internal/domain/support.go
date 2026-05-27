package domain

import "time"

const (
	SupportThreadOpen   = "open"
	SupportThreadClosed = "closed"

	SupportSenderUser  = "user"
	SupportSenderStaff = "staff"
)

type SupportThread struct {
	ID        int64
	UserID    int64
	UserEmail string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SupportMessage struct {
	ID         int64
	ThreadID   int64
	SenderID   int64
	SenderRole string
	SenderName string
	Body       string
	CreatedAt  time.Time
}

type SupportThreadListItem struct {
	SupportThread
	LastMessageBody string
	LastMessageAt   *time.Time
	MessageCount    int
}
