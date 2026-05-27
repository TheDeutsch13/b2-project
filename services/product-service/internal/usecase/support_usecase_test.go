package usecase

import (
	"context"
	"testing"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSupportRepository struct {
	mock.Mock
}

func (m *MockSupportRepository) GetOrCreateThread(
	ctx context.Context,
	userID int64,
	userEmail string,
) (*domain.SupportThread, error) {
	args := m.Called(ctx, userID, userEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SupportThread), args.Error(1)
}

func (m *MockSupportRepository) GetThreadByUserID(
	ctx context.Context,
	userID int64,
) (*domain.SupportThread, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SupportThread), args.Error(1)
}

func (m *MockSupportRepository) GetThreadByID(
	ctx context.Context,
	threadID int64,
) (*domain.SupportThread, error) {
	args := m.Called(ctx, threadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SupportThread), args.Error(1)
}

func (m *MockSupportRepository) ListThreads(
	ctx context.Context,
	openOnly bool,
) ([]domain.SupportThreadListItem, error) {
	args := m.Called(ctx, openOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.SupportThreadListItem), args.Error(1)
}

func (m *MockSupportRepository) UpdateThreadStatus(
	ctx context.Context,
	threadID int64,
	status string,
) (*domain.SupportThread, error) {
	args := m.Called(ctx, threadID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SupportThread), args.Error(1)
}

func (m *MockSupportRepository) DeleteThread(ctx context.Context, threadID int64) error {
	return m.Called(ctx, threadID).Error(0)
}

func (m *MockSupportRepository) ListMessages(
	ctx context.Context,
	threadID int64,
) ([]domain.SupportMessage, error) {
	args := m.Called(ctx, threadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.SupportMessage), args.Error(1)
}

func (m *MockSupportRepository) CreateMessage(
	ctx context.Context,
	threadID int64,
	senderID int64,
	senderRole string,
	senderName string,
	body string,
) (*domain.SupportMessage, error) {
	args := m.Called(ctx, threadID, senderID, senderRole, senderName, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SupportMessage), args.Error(1)
}

func TestSupportUsecase_GetMyThread_Success(t *testing.T) {
	repo := new(MockSupportRepository)
	usecase := NewSupportUsecase(repo)

	thread := &domain.SupportThread{ID: 1, UserID: 5, Status: "open"}
	repo.On("GetOrCreateThread", mock.Anything, int64(5), "u@test.com").Return(thread, nil).Once()
	repo.On("ListMessages", mock.Anything, int64(1)).Return([]domain.SupportMessage{}, nil).Once()

	view, err := usecase.GetMyThread(context.Background(), 5, "u@test.com")

	assert.NoError(t, err)
	assert.Equal(t, int64(1), view.Thread.ID)
}

func TestSupportUsecase_SendUserMessage_EmptyBody(t *testing.T) {
	usecase := NewSupportUsecase(new(MockSupportRepository))

	msg, thread, err := usecase.SendUserMessage(context.Background(), 1, "u@test.com", "User", "  ")

	assert.Nil(t, msg)
	assert.Nil(t, thread)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestSupportUsecase_ListThreadsForStaff_Success(t *testing.T) {
	repo := new(MockSupportRepository)
	usecase := NewSupportUsecase(repo)

	expected := []domain.SupportThreadListItem{{
		SupportThread: domain.SupportThread{ID: 1, UserEmail: "u@test.com"},
	}}
	repo.On("ListThreads", mock.Anything, true).Return(expected, nil).Once()

	threads, err := usecase.ListThreadsForStaff(context.Background(), true)

	assert.NoError(t, err)
	assert.Len(t, threads, 1)
}

func TestSupportUsecase_SendUserMessage_Success(t *testing.T) {
	repo := new(MockSupportRepository)
	usecase := NewSupportUsecase(repo)

	thread := &domain.SupportThread{ID: 2, UserID: 1}
	message := &domain.SupportMessage{ID: 10, Body: "Hello"}

	repo.On("GetOrCreateThread", mock.Anything, int64(1), "u@test.com").Return(thread, nil).Once()
	repo.On("CreateMessage", mock.Anything, int64(2), int64(1), domain.SupportSenderUser, "User", "Hello").
		Return(message, nil).Once()

	msg, gotThread, err := usecase.SendUserMessage(context.Background(), 1, "u@test.com", "User", "Hello")

	assert.NoError(t, err)
	assert.Equal(t, "Hello", msg.Body)
	assert.Equal(t, int64(2), gotThread.ID)
}

func TestSupportUsecase_CloseThread_Success(t *testing.T) {
	repo := new(MockSupportRepository)
	usecase := NewSupportUsecase(repo)

	closed := &domain.SupportThread{ID: 3, Status: domain.SupportThreadClosed}
	repo.On("UpdateThreadStatus", mock.Anything, int64(3), domain.SupportThreadClosed).
		Return(closed, nil).Once()

	thread, err := usecase.CloseThread(context.Background(), 3)

	assert.NoError(t, err)
	assert.Equal(t, domain.SupportThreadClosed, thread.Status)
}

func TestIsSupportStaff(t *testing.T) {
	assert.True(t, IsSupportStaff("admin"))
	assert.True(t, IsSupportStaff("moderator"))
	assert.False(t, IsSupportStaff("user"))
}

func TestSupportUsecase_GetThreadForStaff_Success(t *testing.T) {
	repo := new(MockSupportRepository)
	usecase := NewSupportUsecase(repo)

	thread := &domain.SupportThread{ID: 8, UserID: 2}
	repo.On("GetThreadByID", mock.Anything, int64(8)).Return(thread, nil).Once()
	repo.On("ListMessages", mock.Anything, int64(8)).
		Return([]domain.SupportMessage{{ID: 1, Body: "Question"}}, nil).Once()

	view, err := usecase.GetThreadForStaff(context.Background(), 8)

	assert.NoError(t, err)
	assert.Equal(t, "Question", view.Messages[0].Body)
}

func TestSupportUsecase_SendStaffMessage_Success(t *testing.T) {
	repo := new(MockSupportRepository)
	usecase := NewSupportUsecase(repo)

	thread := &domain.SupportThread{ID: 9, UserID: 2}
	message := &domain.SupportMessage{ID: 11, Body: "Reply"}

	repo.On("GetThreadByID", mock.Anything, int64(9)).Return(thread, nil).Once()
	repo.On(
		"CreateMessage",
		mock.Anything,
		int64(9),
		int64(3),
		domain.SupportSenderStaff,
		"Mod",
		"Reply",
	).Return(message, nil).Once()

	msg, gotThread, err := usecase.SendStaffMessage(context.Background(), 9, 3, "Mod", "Reply")

	assert.NoError(t, err)
	assert.Equal(t, "Reply", msg.Body)
	assert.Equal(t, int64(9), gotThread.ID)
}

func TestSupportUsecase_DeleteThread_Success(t *testing.T) {
	repo := new(MockSupportRepository)
	usecase := NewSupportUsecase(repo)

	repo.On("DeleteThread", mock.Anything, int64(10)).Return(nil).Once()

	err := usecase.DeleteThread(context.Background(), 10)

	assert.NoError(t, err)
}
