package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/jackc/pgx/v5"
)

var (
	ErrSupportThreadNotFound = errors.New("support thread not found")
	ErrSupportAccessDenied   = errors.New("support access denied")
)

type SupportRepository interface {
	GetOrCreateThread(ctx context.Context, userID int64, userEmail string) (*domain.SupportThread, error)
	GetThreadByUserID(ctx context.Context, userID int64) (*domain.SupportThread, error)
	GetThreadByID(ctx context.Context, threadID int64) (*domain.SupportThread, error)
	ListThreads(ctx context.Context, openOnly bool) ([]domain.SupportThreadListItem, error)
	UpdateThreadStatus(ctx context.Context, threadID int64, status string) (*domain.SupportThread, error)
	DeleteThread(ctx context.Context, threadID int64) error
	ListMessages(ctx context.Context, threadID int64) ([]domain.SupportMessage, error)
	CreateMessage(
		ctx context.Context,
		threadID int64,
		senderID int64,
		senderRole string,
		senderName string,
		body string,
	) (*domain.SupportMessage, error)
}

type SupportThreadView struct {
	Thread   domain.SupportThread
	Messages []domain.SupportMessage
}

type SupportUsecase struct {
	supportRepo SupportRepository
}

func NewSupportUsecase(supportRepo SupportRepository) *SupportUsecase {
	return &SupportUsecase{supportRepo: supportRepo}
}

func (u *SupportUsecase) GetMyThread(
	ctx context.Context,
	userID int64,
	userEmail string,
) (*SupportThreadView, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	thread, err := u.supportRepo.GetOrCreateThread(ctx, userID, userEmail)
	if err != nil {
		return nil, err
	}

	messages, err := u.supportRepo.ListMessages(ctx, thread.ID)
	if err != nil {
		return nil, err
	}

	return &SupportThreadView{
		Thread:   *thread,
		Messages: messages,
	}, nil
}

func (u *SupportUsecase) SendUserMessage(
	ctx context.Context,
	userID int64,
	userEmail string,
	senderName string,
	body string,
) (*domain.SupportMessage, *domain.SupportThread, error) {
	body = strings.TrimSpace(body)
	if userID == 0 || body == "" {
		return nil, nil, ErrInvalidInput
	}

	thread, err := u.supportRepo.GetOrCreateThread(ctx, userID, userEmail)
	if err != nil {
		return nil, nil, err
	}

	message, err := u.supportRepo.CreateMessage(
		ctx,
		thread.ID,
		userID,
		domain.SupportSenderUser,
		senderName,
		body,
	)
	if err != nil {
		return nil, nil, err
	}

	return message, thread, nil
}

func (u *SupportUsecase) ListThreadsForStaff(
	ctx context.Context,
	openOnly bool,
) ([]domain.SupportThreadListItem, error) {
	return u.supportRepo.ListThreads(ctx, openOnly)
}

func (u *SupportUsecase) CloseThread(
	ctx context.Context,
	threadID int64,
) (*domain.SupportThread, error) {
	if threadID == 0 {
		return nil, ErrInvalidInput
	}

	thread, err := u.supportRepo.UpdateThreadStatus(ctx, threadID, domain.SupportThreadClosed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSupportThreadNotFound
		}
		return nil, err
	}

	return thread, nil
}

func (u *SupportUsecase) DeleteThread(ctx context.Context, threadID int64) error {
	if threadID == 0 {
		return ErrInvalidInput
	}

	err := u.supportRepo.DeleteThread(ctx, threadID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSupportThreadNotFound
		}
		return err
	}

	return nil
}

func (u *SupportUsecase) GetThreadForStaff(
	ctx context.Context,
	threadID int64,
) (*SupportThreadView, error) {
	if threadID == 0 {
		return nil, ErrInvalidInput
	}

	thread, err := u.supportRepo.GetThreadByID(ctx, threadID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSupportThreadNotFound
		}
		return nil, err
	}

	messages, err := u.supportRepo.ListMessages(ctx, thread.ID)
	if err != nil {
		return nil, err
	}

	return &SupportThreadView{
		Thread:   *thread,
		Messages: messages,
	}, nil
}

func (u *SupportUsecase) SendStaffMessage(
	ctx context.Context,
	threadID int64,
	staffID int64,
	senderName string,
	body string,
) (*domain.SupportMessage, *domain.SupportThread, error) {
	body = strings.TrimSpace(body)
	if threadID == 0 || staffID == 0 || body == "" {
		return nil, nil, ErrInvalidInput
	}

	thread, err := u.supportRepo.GetThreadByID(ctx, threadID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrSupportThreadNotFound
		}
		return nil, nil, err
	}

	message, err := u.supportRepo.CreateMessage(
		ctx,
		thread.ID,
		staffID,
		domain.SupportSenderStaff,
		senderName,
		body,
	)
	if err != nil {
		return nil, nil, err
	}

	return message, thread, nil
}

func IsSupportStaff(role string) bool {
	return role == "admin" || role == "moderator"
}
