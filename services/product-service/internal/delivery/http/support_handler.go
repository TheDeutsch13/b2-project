package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/TheDeutsch13/b2-common/httperr"
	commonmiddleware "github.com/TheDeutsch13/b2-common/middleware"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/delivery/ws"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SupportHandler struct {
	supportUsecase *usecase.SupportUsecase
	hub            *ws.Hub
	logger         *zap.Logger
}

func NewSupportHandler(
	supportUsecase *usecase.SupportUsecase,
	hub *ws.Hub,
	logger *zap.Logger,
) *SupportHandler {
	return &SupportHandler{
		supportUsecase: supportUsecase,
		hub:            hub,
		logger:         logger,
	}
}

type supportMessageRequest struct {
	Body       string `json:"body"`
	SenderName string `json:"sender_name"`
}

type supportMessageResponse struct {
	ID         int64     `json:"id"`
	ThreadID   int64     `json:"thread_id"`
	SenderID   int64     `json:"sender_id"`
	SenderRole string    `json:"sender_role"`
	SenderName string    `json:"sender_name"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}

type supportThreadResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	UserEmail string    `json:"user_email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type supportThreadListItemResponse struct {
	supportThreadResponse
	LastMessageBody string     `json:"last_message_body"`
	LastMessageAt   *time.Time `json:"last_message_at,omitempty"`
	MessageCount    int        `json:"message_count"`
}

type supportThreadViewResponse struct {
	Thread   supportThreadResponse    `json:"thread"`
	Messages []supportMessageResponse `json:"messages"`
}

func toSupportMessageResponse(message domain.SupportMessage) supportMessageResponse {
	return supportMessageResponse{
		ID:         message.ID,
		ThreadID:   message.ThreadID,
		SenderID:   message.SenderID,
		SenderRole: message.SenderRole,
		SenderName: message.SenderName,
		Body:       message.Body,
		CreatedAt:  message.CreatedAt,
	}
}

func toSupportThreadResponse(thread domain.SupportThread) supportThreadResponse {
	return supportThreadResponse{
		ID:        thread.ID,
		UserID:    thread.UserID,
		UserEmail: thread.UserEmail,
		Status:    thread.Status,
		CreatedAt: thread.CreatedAt,
		UpdatedAt: thread.UpdatedAt,
	}
}

func toSupportThreadViewResponse(view *usecase.SupportThreadView) supportThreadViewResponse {
	messages := make([]supportMessageResponse, 0, len(view.Messages))
	for index := range view.Messages {
		messages = append(messages, toSupportMessageResponse(view.Messages[index]))
	}

	return supportThreadViewResponse{
		Thread:   toSupportThreadResponse(view.Thread),
		Messages: messages,
	}
}

func (h *SupportHandler) broadcastSupportMessage(
	message domain.SupportMessage,
	thread domain.SupportThread,
) {
	h.hub.Broadcast(ws.Notification{
		Type:              "support_message",
		Message:           "Новое сообщение в поддержке",
		SupportThreadID:   thread.ID,
		SupportMessageID:  message.ID,
		SupportBody:       message.Body,
		SupportSenderID:   message.SenderID,
		SupportSenderRole: message.SenderRole,
		SupportSenderName: message.SenderName,
		TargetUserID:      thread.UserID,
	})
}

func (h *SupportHandler) GetMyThread(ctx *gin.Context) {
	userID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	email, _ := ctx.Get(commonmiddleware.ContextEmailKey)
	userEmail, _ := email.(string)

	view, err := h.supportUsecase.GetMyThread(ctx.Request.Context(), userID, userEmail)
	if err != nil {
		h.logger.Error("failed to get support thread", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusOK, toSupportThreadViewResponse(view))
}

func (h *SupportHandler) SendMyMessage(ctx *gin.Context) {
	userID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	var req supportMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	email, _ := ctx.Get(commonmiddleware.ContextEmailKey)
	userEmail, _ := email.(string)

	message, thread, err := h.supportUsecase.SendUserMessage(
		ctx.Request.Context(),
		userID,
		userEmail,
		req.SenderName,
		req.Body,
	)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			httperr.BadRequest(ctx, "invalid message")
			return
		}

		h.logger.Error("failed to send support message", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	h.broadcastSupportMessage(*message, *thread)
	ctx.JSON(stdhttp.StatusCreated, toSupportMessageResponse(*message))
}

func (h *SupportHandler) ListThreads(ctx *gin.Context) {
	includeClosed :=
		ctx.Query("include_closed") == "true" || ctx.Query("include_closed") == "1"

	threads, err := h.supportUsecase.ListThreadsForStaff(
		ctx.Request.Context(),
		!includeClosed,
	)
	if err != nil {
		h.logger.Error("failed to list support threads", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	response := make([]supportThreadListItemResponse, 0, len(threads))
	for index := range threads {
		item := threads[index]
		response = append(response, supportThreadListItemResponse{
			supportThreadResponse: toSupportThreadResponse(item.SupportThread),
			LastMessageBody:       item.LastMessageBody,
			LastMessageAt:         item.LastMessageAt,
			MessageCount:          item.MessageCount,
		})
	}

	ctx.JSON(stdhttp.StatusOK, response)
}

func (h *SupportHandler) GetThread(ctx *gin.Context) {
	threadID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || threadID <= 0 {
		httperr.BadRequest(ctx, "invalid thread id")
		return
	}

	view, err := h.supportUsecase.GetThreadForStaff(ctx.Request.Context(), threadID)
	if err != nil {
		if errors.Is(err, usecase.ErrSupportThreadNotFound) {
			httperr.BadRequest(ctx, "thread not found")
			return
		}

		h.logger.Error("failed to get support thread", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusOK, toSupportThreadViewResponse(view))
}

func (h *SupportHandler) SendThreadMessage(ctx *gin.Context) {
	staffID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	threadID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || threadID <= 0 {
		httperr.BadRequest(ctx, "invalid thread id")
		return
	}

	var req supportMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	message, thread, err := h.supportUsecase.SendStaffMessage(
		ctx.Request.Context(),
		threadID,
		staffID,
		req.SenderName,
		req.Body,
	)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			httperr.BadRequest(ctx, "invalid message")
			return
		}
		if errors.Is(err, usecase.ErrSupportThreadNotFound) {
			httperr.BadRequest(ctx, "thread not found")
			return
		}

		h.logger.Error("failed to send staff support message", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	h.broadcastSupportMessage(*message, *thread)
	ctx.JSON(stdhttp.StatusCreated, toSupportMessageResponse(*message))
}

func (h *SupportHandler) CloseThread(ctx *gin.Context) {
	threadID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || threadID <= 0 {
		httperr.BadRequest(ctx, "invalid thread id")
		return
	}

	thread, err := h.supportUsecase.CloseThread(ctx.Request.Context(), threadID)
	if err != nil {
		if errors.Is(err, usecase.ErrSupportThreadNotFound) {
			httperr.BadRequest(ctx, "thread not found")
			return
		}

		h.logger.Error("failed to close support thread", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusOK, toSupportThreadResponse(*thread))
}

func (h *SupportHandler) DeleteThread(ctx *gin.Context) {
	threadID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || threadID <= 0 {
		httperr.BadRequest(ctx, "invalid thread id")
		return
	}

	if err := h.supportUsecase.DeleteThread(ctx.Request.Context(), threadID); err != nil {
		if errors.Is(err, usecase.ErrSupportThreadNotFound) {
			httperr.BadRequest(ctx, "thread not found")
			return
		}

		h.logger.Error("failed to delete support thread", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.Status(stdhttp.StatusNoContent)
}
