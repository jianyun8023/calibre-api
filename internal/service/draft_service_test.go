package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jianyun8023/calibre-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockDraftRepository struct {
	mock.Mock
}

func (m *MockDraftRepository) CreateDraft(ctx context.Context, draft *repository.BookDraft) (int64, error) {
	args := m.Called(ctx, draft)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDraftRepository) GetPendingDrafts(ctx context.Context, limit, offset int) ([]repository.BookDraft, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]repository.BookDraft), args.Get(1).(int64), args.Error(2)
}

func (m *MockDraftRepository) GetDraftByID(ctx context.Context, id int64) (*repository.BookDraft, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.BookDraft), args.Error(1)
}

func (m *MockDraftRepository) GetPendingDraftByBookIDAndAction(ctx context.Context, bookID string, action repository.DraftType) (*repository.BookDraft, error) {
	args := m.Called(ctx, bookID, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.BookDraft), args.Error(1)
}

func (m *MockDraftRepository) GetPendingDraftsByBookIDsAndAction(ctx context.Context, bookIDs []string, action repository.DraftType) ([]repository.BookDraft, error) {
	args := m.Called(ctx, bookIDs, action)
	return args.Get(0).([]repository.BookDraft), args.Error(1)
}

func (m *MockDraftRepository) GetPendingDraftsBefore(ctx context.Context, before time.Time) ([]repository.BookDraft, error) {
	args := m.Called(ctx, mock.Anything)
	return args.Get(0).([]repository.BookDraft), args.Error(1)
}

func (m *MockDraftRepository) ResetStuckProcessingDrafts(ctx context.Context, threshold time.Time) (int64, error) {
	args := m.Called(ctx, mock.Anything)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDraftRepository) UpdateDraftStatus(ctx context.Context, id int64, status repository.DraftStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockDraftRepository) UpdateDraftData(ctx context.Context, id int64, data string) error {
	args := m.Called(ctx, id, data)
	return args.Error(0)
}

func (m *MockDraftRepository) DeleteDraft(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDraftRepository) UpdateDraftStatusIfPending(ctx context.Context, id int64, newStatus repository.DraftStatus) (bool, error) {
	args := m.Called(ctx, id, newStatus)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockDraftRepository) ApplyDraftSuccess(ctx context.Context, draft *repository.BookDraft, newStatus repository.DraftStatus) error {
	args := m.Called(ctx, draft, newStatus)
	return args.Error(0)
}

func (m *MockDraftRepository) ExpireDraftAtomically(ctx context.Context, draft *repository.BookDraft) error {
	args := m.Called(ctx, draft)
	return args.Error(0)
}

func (m *MockDraftRepository) CreateHistory(ctx context.Context, history *repository.BookDraftHistory) (int64, error) {
	args := m.Called(ctx, history)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDraftRepository) GetHistory(ctx context.Context, limit, offset int) ([]repository.BookDraftHistory, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]repository.BookDraftHistory), args.Get(1).(int64), args.Error(2)
}

// Mock BookService
type MockBookService struct {
	mock.Mock
}

func (m *MockBookService) GetBookByID(id string) (*Book, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Book), args.Error(1)
}

func (m *MockBookService) DeleteBook(bookID string) error {
	args := m.Called(bookID)
	return args.Error(0)
}

func (m *MockBookService) UpdateMetadata(bookID string, metadata *BookUpdate) error {
	args := m.Called(bookID, metadata)
	return args.Error(0)
}

func (m *MockBookService) GetRecentBooks(limit, offset int) ([]Book, int64, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]Book), args.Get(1).(int64), args.Error(2)
}

func (m *MockBookService) GetRandomBooks(limit int) ([]Book, error) {
	args := m.Called(limit)
	return args.Get(0).([]Book), args.Error(1)
}

func (m *MockBookService) GetAllBooks(limit int, cursor string) ([]Book, int64, string, error) {
	args := m.Called(limit, cursor)
	return args.Get(0).([]Book), args.Get(1).(int64), args.Get(2).(string), args.Error(3)
}

func (m *MockBookService) ListPublishers() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

// Tests

func TestReceiveDeletes_Success(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	ids := []string{"1", "2", "3"}

	// Mock: 没有已存在的草稿
	mockRepo.On("GetPendingDraftsByBookIDsAndAction", ctx, ids, repository.DraftActionDelete).
		Return([]repository.BookDraft{}, nil)

	// Mock: 创建草稿
	mockRepo.On("CreateDraft", ctx, mock.Anything).Return(int64(1), nil).Times(3)

	err := service.ReceiveDeletes(ctx, ids)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestReceiveDeletes_SkipDuplicates(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	ids := []string{"1", "2"}

	// Mock: 书籍1已有待审核的删除草稿
	existingDrafts := []repository.BookDraft{
		{ID: 1, BookID: "1", Action: repository.DraftActionDelete, Status: repository.DraftStatusPending},
	}
	mockRepo.On("GetPendingDraftsByBookIDsAndAction", ctx, ids, repository.DraftActionDelete).
		Return(existingDrafts, nil)

	// Mock: 只为书籍2创建草稿
	mockRepo.On("CreateDraft", ctx, mock.MatchedBy(func(draft *repository.BookDraft) bool {
		return draft.BookID == "2"
	})).Return(int64(2), nil)

	err := service.ReceiveDeletes(ctx, ids)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "CreateDraft", 1)
}

func TestReceiveUpdates_Success(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	title := "New Title"
	updates := []BookDraftUpdate{
		{
			ID:   "1",
			Data: &BookUpdate{Title: &title},
		},
	}

	// Mock: 没有已存在的草稿
	mockRepo.On("GetPendingDraftsByBookIDsAndAction", ctx, []string{"1"}, repository.DraftActionUpdate).
		Return([]repository.BookDraft{}, nil)

	// Mock: 创建草稿
	mockRepo.On("CreateDraft", ctx, mock.Anything).Return(int64(1), nil)

	err := service.ReceiveUpdates(ctx, updates)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestReceiveUpdates_UpdateExisting(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	title := "Updated Title"
	updates := []BookDraftUpdate{
		{
			ID:   "1",
			Data: &BookUpdate{Title: &title},
		},
	}

	// Mock: 已存在待审核的更新草稿
	existingDrafts := []repository.BookDraft{
		{ID: 1, BookID: "1", Action: repository.DraftActionUpdate, Status: repository.DraftStatusPending},
	}
	mockRepo.On("GetPendingDraftsByBookIDsAndAction", ctx, []string{"1"}, repository.DraftActionUpdate).
		Return(existingDrafts, nil)

	// Mock: 更新现有草稿的数据
	mockRepo.On("UpdateDraftData", ctx, int64(1), mock.Anything).Return(nil)

	err := service.ReceiveUpdates(ctx, updates)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "CreateDraft")
}

func TestGetPendingDrafts_Success(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	expectedDrafts := []repository.BookDraft{
		{ID: 1, BookID: "1", Action: repository.DraftActionDelete, Status: repository.DraftStatusPending},
		{ID: 2, BookID: "2", Action: repository.DraftActionUpdate, Status: repository.DraftStatusPending},
	}

	mockRepo.On("GetPendingDrafts", ctx, 10, 0).Return(expectedDrafts, int64(2), nil)

	drafts, total, err := service.GetPendingDrafts(ctx, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedDrafts, drafts)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestApplyDrafts_DeleteSuccess(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	draftID := int64(1)
	draft := &repository.BookDraft{
		ID:     draftID,
		BookID: "123",
		Action: repository.DraftActionDelete,
		Status: repository.DraftStatusPending,
	}

	// Mock 流程
	mockRepo.On("GetDraftByID", ctx, draftID).Return(draft, nil)
	mockRepo.On("UpdateDraftStatusIfPending", ctx, draftID, repository.DraftStatusProcessing).Return(true, nil)
	mockBookService.On("DeleteBook", "123").Return(nil)
	mockRepo.On("ApplyDraftSuccess", ctx, draft, repository.DraftStatusApplied).Return(nil)

	errs := service.ApplyDrafts(ctx, []int64{draftID})

	assert.Empty(t, errs)
	mockRepo.AssertExpectations(t)
	mockBookService.AssertExpectations(t)
}

func TestApplyDrafts_UpdateSuccess(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	draftID := int64(1)

	title := "New Title"
	updateData := &BookUpdate{Title: &title}
	dataJSON, _ := json.Marshal(updateData)

	draft := &repository.BookDraft{
		ID:     draftID,
		BookID: "123",
		Action: repository.DraftActionUpdate,
		Data:   string(dataJSON),
		Status: repository.DraftStatusPending,
	}

	// Mock 流程
	mockRepo.On("GetDraftByID", ctx, draftID).Return(draft, nil)
	mockRepo.On("UpdateDraftStatusIfPending", ctx, draftID, repository.DraftStatusProcessing).Return(true, nil)
	mockBookService.On("UpdateMetadata", "123", mock.Anything).Return(nil)
	mockRepo.On("ApplyDraftSuccess", ctx, draft, repository.DraftStatusApplied).Return(nil)

	errs := service.ApplyDrafts(ctx, []int64{draftID})

	assert.Empty(t, errs)
	mockRepo.AssertExpectations(t)
	mockBookService.AssertExpectations(t)
}

func TestApplyDrafts_ExternalActionFailure(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	draftID := int64(1)
	draft := &repository.BookDraft{
		ID:     draftID,
		BookID: "123",
		Action: repository.DraftActionDelete,
		Status: repository.DraftStatusPending,
	}

	// Mock 流程：外部操作失败
	mockRepo.On("GetDraftByID", ctx, draftID).Return(draft, nil)
	mockRepo.On("UpdateDraftStatusIfPending", ctx, draftID, repository.DraftStatusProcessing).Return(true, nil)
	mockBookService.On("DeleteBook", "123").Return(errors.New("external service error"))
	mockRepo.On("UpdateDraftStatus", ctx, draftID, repository.DraftStatusPending).Return(nil)

	errs := service.ApplyDrafts(ctx, []int64{draftID})

	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "external service error")
	mockRepo.AssertExpectations(t)
	mockBookService.AssertExpectations(t)
}

func TestApplyDrafts_ConcurrentProcessing(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	draftID := int64(1)
	draft := &repository.BookDraft{
		ID:     draftID,
		BookID: "123",
		Action: repository.DraftActionDelete,
		Status: repository.DraftStatusPending,
	}

	// Mock: 已被其他进程标记为 processing
	mockRepo.On("GetDraftByID", ctx, draftID).Return(draft, nil)
	mockRepo.On("UpdateDraftStatusIfPending", ctx, draftID, repository.DraftStatusProcessing).Return(false, nil)

	errs := service.ApplyDrafts(ctx, []int64{draftID})

	// 应该没有错误，因为已被其他进程处理
	assert.Empty(t, errs)
	mockRepo.AssertExpectations(t)
	mockBookService.AssertNotCalled(t, "DeleteBook")
}

func TestRejectDrafts_Success(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	draftID := int64(1)
	draft := &repository.BookDraft{
		ID:     draftID,
		BookID: "123",
		Action: repository.DraftActionDelete,
		Status: repository.DraftStatusPending,
	}

	mockRepo.On("GetDraftByID", ctx, draftID).Return(draft, nil)
	mockRepo.On("UpdateDraftStatusIfPending", ctx, draftID, repository.DraftStatusProcessing).Return(true, nil)
	mockRepo.On("ApplyDraftSuccess", ctx, draft, repository.DraftStatusRejected).Return(nil)

	errs := service.RejectDrafts(ctx, []int64{draftID})

	assert.Empty(t, errs)
	mockRepo.AssertExpectations(t)
}

func TestCleanupExpiredDrafts_Success(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()

	// Mock: 重置卡住的草稿
	mockRepo.On("ResetStuckProcessingDrafts", ctx, mock.Anything).Return(int64(2), nil)

	// Mock: 获取过期草稿
	expiredDrafts := []repository.BookDraft{
		{ID: 1, BookID: "1", Action: repository.DraftActionDelete, Status: repository.DraftStatusPending},
		{ID: 2, BookID: "2", Action: repository.DraftActionUpdate, Status: repository.DraftStatusPending},
	}
	mockRepo.On("GetPendingDraftsBefore", ctx, mock.Anything).Return(expiredDrafts, nil)

	// Mock: 过期草稿
	mockRepo.On("ExpireDraftAtomically", ctx, mock.Anything).Return(nil).Times(2)

	count, err := service.CleanupExpiredDrafts(ctx, 30)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	mockRepo.AssertExpectations(t)
}

func TestGetHistory_Success(t *testing.T) {
	mockRepo := new(MockDraftRepository)
	mockBookService := new(MockBookService)
	service := NewDraftService(mockRepo, mockBookService)

	ctx := context.Background()
	expectedHistory := []repository.BookDraftHistory{
		{ID: 1, DraftID: 1, BookID: "1", Status: repository.DraftStatusApplied},
		{ID: 2, DraftID: 2, BookID: "2", Status: repository.DraftStatusRejected},
	}

	mockRepo.On("GetHistory", ctx, 50, 0).Return(expectedHistory, int64(2), nil)

	histories, total, err := service.GetHistory(ctx, 50, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedHistory, histories)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}
