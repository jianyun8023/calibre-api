package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jianyun8023/calibre-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDraftRepositoryCancel implements DraftRepository for CancelDrafts tests.
type MockDraftRepositoryCancel struct {
	mock.Mock
}

func (m *MockDraftRepositoryCancel) GetPendingDraftByBookIDAndAction(ctx context.Context, bookID string, action repository.DraftType) (*repository.BookDraft, error) {
	args := m.Called(ctx, bookID, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.BookDraft), args.Error(1)
}

func (m *MockDraftRepositoryCancel) DeleteDraft(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDraftRepositoryCancel) CreateDraft(ctx context.Context, draft *repository.BookDraft) (int64, error) {
	return 0, nil
}

func (m *MockDraftRepositoryCancel) GetPendingDrafts(ctx context.Context, limit, offset int) ([]repository.BookDraft, int64, error) {
	return nil, 0, nil
}

func (m *MockDraftRepositoryCancel) GetDraftByID(ctx context.Context, id int64) (*repository.BookDraft, error) {
	return nil, nil
}

func (m *MockDraftRepositoryCancel) GetPendingDraftsByBookIDsAndAction(ctx context.Context, bookIDs []string, action repository.DraftType) ([]repository.BookDraft, error) {
	return nil, nil
}

func (m *MockDraftRepositoryCancel) GetPendingDraftsBefore(ctx context.Context, before time.Time) ([]repository.BookDraft, error) {
	return nil, nil
}

func (m *MockDraftRepositoryCancel) ResetStuckProcessingDrafts(ctx context.Context, threshold time.Time) (int64, error) {
	return 0, nil
}

func (m *MockDraftRepositoryCancel) UpdateDraftStatus(ctx context.Context, id int64, status repository.DraftStatus) error {
	return nil
}

func (m *MockDraftRepositoryCancel) UpdateDraftData(ctx context.Context, id int64, data string) error {
	return nil
}

func (m *MockDraftRepositoryCancel) UpdateDraftStatusIfPending(ctx context.Context, id int64, newStatus repository.DraftStatus) (bool, error) {
	return false, nil
}

func (m *MockDraftRepositoryCancel) ApplyDraftSuccess(ctx context.Context, draft *repository.BookDraft, newStatus repository.DraftStatus) error {
	return nil
}

func (m *MockDraftRepositoryCancel) ExpireDraftAtomically(ctx context.Context, draft *repository.BookDraft) error {
	return nil
}

func (m *MockDraftRepositoryCancel) CreateHistory(ctx context.Context, history *repository.BookDraftHistory) (int64, error) {
	return 0, nil
}

func (m *MockDraftRepositoryCancel) GetHistory(ctx context.Context, limit, offset int) ([]repository.BookDraftHistory, int64, error) {
	return nil, 0, nil
}

func TestCancelDrafts_SingleBookWithOneDraft(t *testing.T) {
	mockRepo := new(MockDraftRepositoryCancel)
	mockBookService := new(MockBookService)
	svc := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	bookID := "274752"

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID, repository.DraftActionUpdate).
		Return(&repository.BookDraft{ID: 1, BookID: bookID, Action: repository.DraftActionUpdate, Status: repository.DraftStatusPending}, nil)

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID, repository.DraftActionDelete).
		Return(nil, nil)

	mockRepo.On("DeleteDraft", ctx, int64(1)).Return(nil)

	cancelledBooks, cancelledDrafts, errs := svc.CancelDrafts(ctx, []string{bookID})

	assert.Equal(t, 1, cancelledBooks)
	assert.Equal(t, 1, cancelledDrafts)
	assert.Empty(t, errs)
	mockRepo.AssertExpectations(t)
}

func TestCancelDrafts_SingleBookWithMultipleDrafts(t *testing.T) {
	mockRepo := new(MockDraftRepositoryCancel)
	mockBookService := new(MockBookService)
	svc := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	bookID := "274752"

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID, repository.DraftActionUpdate).
		Return(&repository.BookDraft{ID: 1, BookID: bookID, Action: repository.DraftActionUpdate, Status: repository.DraftStatusPending}, nil)

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID, repository.DraftActionDelete).
		Return(&repository.BookDraft{ID: 2, BookID: bookID, Action: repository.DraftActionDelete, Status: repository.DraftStatusPending}, nil)

	mockRepo.On("DeleteDraft", ctx, int64(1)).Return(nil)
	mockRepo.On("DeleteDraft", ctx, int64(2)).Return(nil)

	cancelledBooks, cancelledDrafts, errs := svc.CancelDrafts(ctx, []string{bookID})

	assert.Equal(t, 1, cancelledBooks)
	assert.Equal(t, 2, cancelledDrafts)
	assert.Empty(t, errs)
	mockRepo.AssertExpectations(t)
}

func TestCancelDrafts_BookWithNoDrafts(t *testing.T) {
	mockRepo := new(MockDraftRepositoryCancel)
	mockBookService := new(MockBookService)
	svc := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	bookID := "999999"

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID, repository.DraftActionUpdate).
		Return(nil, nil)

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID, repository.DraftActionDelete).
		Return(nil, nil)

	cancelledBooks, cancelledDrafts, errs := svc.CancelDrafts(ctx, []string{bookID})

	assert.Equal(t, 0, cancelledBooks)
	assert.Equal(t, 0, cancelledDrafts)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "no pending drafts found for book 999999")
	mockRepo.AssertExpectations(t)
}

func TestCancelDrafts_MultipleBooks(t *testing.T) {
	mockRepo := new(MockDraftRepositoryCancel)
	mockBookService := new(MockBookService)
	svc := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	bookID1 := "274752"
	bookID2 := "274781"

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID1, repository.DraftActionUpdate).
		Return(&repository.BookDraft{ID: 1, BookID: bookID1, Action: repository.DraftActionUpdate, Status: repository.DraftStatusPending}, nil)
	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID1, repository.DraftActionDelete).
		Return(nil, nil)
	mockRepo.On("DeleteDraft", ctx, int64(1)).Return(nil)

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID2, repository.DraftActionUpdate).
		Return(&repository.BookDraft{ID: 2, BookID: bookID2, Action: repository.DraftActionUpdate, Status: repository.DraftStatusPending}, nil)
	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID2, repository.DraftActionDelete).
		Return(nil, nil)
	mockRepo.On("DeleteDraft", ctx, int64(2)).Return(nil)

	cancelledBooks, cancelledDrafts, errs := svc.CancelDrafts(ctx, []string{bookID1, bookID2})

	assert.Equal(t, 2, cancelledBooks)
	assert.Equal(t, 2, cancelledDrafts)
	assert.Empty(t, errs)
	mockRepo.AssertExpectations(t)
}

func TestCancelDrafts_DeleteFailure(t *testing.T) {
	mockRepo := new(MockDraftRepositoryCancel)
	mockBookService := new(MockBookService)
	svc := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	bookID := "274752"

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID, repository.DraftActionUpdate).
		Return(&repository.BookDraft{ID: 1, BookID: bookID, Action: repository.DraftActionUpdate, Status: repository.DraftStatusPending}, nil)

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID, repository.DraftActionDelete).
		Return(nil, nil)

	mockRepo.On("DeleteDraft", ctx, int64(1)).Return(fmt.Errorf("database error"))

	cancelledBooks, cancelledDrafts, errs := svc.CancelDrafts(ctx, []string{bookID})

	assert.Equal(t, 0, cancelledBooks)
	assert.Equal(t, 0, cancelledDrafts)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "failed to delete draft 1")
	mockRepo.AssertExpectations(t)
}

func TestCancelDrafts_PartialSuccess(t *testing.T) {
	mockRepo := new(MockDraftRepositoryCancel)
	mockBookService := new(MockBookService)
	svc := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	bookID1 := "274752"
	bookID2 := "999999"
	bookID3 := "274781"

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID1, repository.DraftActionUpdate).
		Return(&repository.BookDraft{ID: 1, BookID: bookID1, Action: repository.DraftActionUpdate, Status: repository.DraftStatusPending}, nil)
	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID1, repository.DraftActionDelete).
		Return(nil, nil)
	mockRepo.On("DeleteDraft", ctx, int64(1)).Return(nil)

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID2, repository.DraftActionUpdate).
		Return(nil, nil)
	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID2, repository.DraftActionDelete).
		Return(nil, nil)

	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID3, repository.DraftActionUpdate).
		Return(&repository.BookDraft{ID: 2, BookID: bookID3, Action: repository.DraftActionUpdate, Status: repository.DraftStatusPending}, nil)
	mockRepo.On("GetPendingDraftByBookIDAndAction", ctx, bookID3, repository.DraftActionDelete).
		Return(nil, nil)
	mockRepo.On("DeleteDraft", ctx, int64(2)).Return(nil)

	cancelledBooks, cancelledDrafts, errs := svc.CancelDrafts(ctx, []string{bookID1, bookID2, bookID3})

	assert.Equal(t, 2, cancelledBooks)
	assert.Equal(t, 2, cancelledDrafts)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "no pending drafts found for book 999999")
	mockRepo.AssertExpectations(t)
}
