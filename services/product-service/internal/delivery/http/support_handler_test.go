package http

import (
	"bytes"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSupportHandler_GetMyThread_Success(t *testing.T) {
	supportRepo := new(mockSupportRepository)
	router := setupTestRouterFull(routerTestDeps{supportRepo: supportRepo})

	thread := &domain.SupportThread{
		ID:        1,
		UserID:    1,
		UserEmail: "user@example.com",
		Status:    domain.SupportThreadOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	supportRepo.On("GetOrCreateThread", mock.Anything, int64(1), "user@example.com").
		Return(thread, nil).Once()
	supportRepo.On("ListMessages", mock.Anything, int64(1)).
		Return([]domain.SupportMessage{}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/support/my", nil)
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":1`)
}

func TestSupportHandler_SendMyMessage_Success(t *testing.T) {
	supportRepo := new(mockSupportRepository)
	router := setupTestRouterFull(routerTestDeps{supportRepo: supportRepo})

	thread := &domain.SupportThread{ID: 2, UserID: 1, UserEmail: "user@example.com"}
	message := &domain.SupportMessage{
		ID:         99,
		ThreadID:   2,
		SenderID:   1,
		SenderRole: domain.SupportSenderUser,
		SenderName: "User",
		Body:       "Need help",
		CreatedAt:  time.Now(),
	}

	supportRepo.On("GetOrCreateThread", mock.Anything, int64(1), "user@example.com").
		Return(thread, nil).Once()
	supportRepo.On(
		"CreateMessage",
		mock.Anything,
		int64(2),
		int64(1),
		domain.SupportSenderUser,
		"User",
		"Need help",
	).Return(message, nil).Once()

	body := []byte(`{"body":"Need help","sender_name":"User"}`)
	req := httptest.NewRequest(
		stdhttp.MethodPost,
		"/api/support/my/messages",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "Need help")
}

func TestSupportHandler_ListThreads_Moderator(t *testing.T) {
	supportRepo := new(mockSupportRepository)
	router := setupTestRouterFull(routerTestDeps{supportRepo: supportRepo})

	supportRepo.On("ListThreads", mock.Anything, true).
		Return([]domain.SupportThreadListItem{{
			SupportThread: domain.SupportThread{
				ID:        3,
				UserEmail: "client@example.com",
				Status:    domain.SupportThreadOpen,
			},
			LastMessageBody: "Hi",
			MessageCount:    1,
		}}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/support/threads", nil)
	req.Header.Set("Authorization", bearerToken(2, "mod@example.com", "moderator"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "client@example.com")
}

func TestSupportHandler_CloseThread_Admin(t *testing.T) {
	supportRepo := new(mockSupportRepository)
	router := setupTestRouterFull(routerTestDeps{supportRepo: supportRepo})

	closed := &domain.SupportThread{
		ID:     4,
		Status: domain.SupportThreadClosed,
	}
	supportRepo.On("UpdateThreadStatus", mock.Anything, int64(4), domain.SupportThreadClosed).
		Return(closed, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodPatch, "/api/support/threads/4/close", nil)
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
}

func TestSupportHandler_GetThread_Staff(t *testing.T) {
	supportRepo := new(mockSupportRepository)
	router := setupTestRouterFull(routerTestDeps{supportRepo: supportRepo})

	thread := &domain.SupportThread{ID: 5, UserID: 2, UserEmail: "client@example.com"}
	supportRepo.On("GetThreadByID", mock.Anything, int64(5)).Return(thread, nil).Once()
	supportRepo.On("ListMessages", mock.Anything, int64(5)).
		Return([]domain.SupportMessage{{
			ID: 1, ThreadID: 5, Body: "Hi", SenderRole: domain.SupportSenderUser,
		}}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/support/threads/5", nil)
	req.Header.Set("Authorization", bearerToken(2, "mod@example.com", "moderator"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Hi")
}

func TestSupportHandler_SendThreadMessage_Staff(t *testing.T) {
	supportRepo := new(mockSupportRepository)
	router := setupTestRouterFull(routerTestDeps{supportRepo: supportRepo})

	thread := &domain.SupportThread{ID: 6, UserID: 2}
	message := &domain.SupportMessage{
		ID: 20, ThreadID: 6, SenderID: 3, SenderRole: domain.SupportSenderStaff,
		Body: "We will help",
	}

	supportRepo.On("GetThreadByID", mock.Anything, int64(6)).Return(thread, nil).Once()
	supportRepo.On(
		"CreateMessage",
		mock.Anything,
		int64(6),
		int64(3),
		domain.SupportSenderStaff,
		"Support",
		"We will help",
	).Return(message, nil).Once()

	body := []byte(`{"body":"We will help","sender_name":"Support"}`)
	req := httptest.NewRequest(
		stdhttp.MethodPost,
		"/api/support/threads/6/messages",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(3, "mod@example.com", "moderator"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusCreated, rec.Code)
}

func TestSupportHandler_DeleteThread_Admin(t *testing.T) {
	supportRepo := new(mockSupportRepository)
	router := setupTestRouterFull(routerTestDeps{supportRepo: supportRepo})

	supportRepo.On("DeleteThread", mock.Anything, int64(7)).Return(nil).Once()

	req := httptest.NewRequest(stdhttp.MethodDelete, "/api/support/threads/7", nil)
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusNoContent, rec.Code)
}

func TestSupportHandler_SendMyMessage_InvalidBody(t *testing.T) {
	supportRepo := new(mockSupportRepository)
	router := setupTestRouterFull(routerTestDeps{supportRepo: supportRepo})

	body := []byte(`{"body":"  "}`)
	req := httptest.NewRequest(
		stdhttp.MethodPost,
		"/api/support/my/messages",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusBadRequest, rec.Code)
}
