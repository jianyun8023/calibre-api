package calibre

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/repository"
	"github.com/jianyun8023/calibre-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDraftServiceCancel struct {
	mock.Mock
}

func (m *MockDraftServiceCancel) CancelDrafts(ctx context.Context, bookIDs []string) (int, int, []error) {
	args := m.Called(ctx, bookIDs)
	var errs []error
	if v := args.Get(2); v != nil {
		errs = v.([]error)
	}
	return args.Int(0), args.Int(1), errs
}

func (m *MockDraftServiceCancel) ReceiveDeletes(ctx context.Context, ids []string) error { return nil }

func (m *MockDraftServiceCancel) ReceiveUpdates(ctx context.Context, updates []service.BookDraftUpdate) error {
	return nil
}

func (m *MockDraftServiceCancel) GetPendingDrafts(ctx context.Context, limit, offset int) ([]repository.BookDraft, int64, error) {
	return nil, 0, nil
}

func (m *MockDraftServiceCancel) ApplyDrafts(ctx context.Context, ids []int64) []error { return nil }

func (m *MockDraftServiceCancel) RejectDrafts(ctx context.Context, ids []int64) []error { return nil }

func (m *MockDraftServiceCancel) CleanupExpiredDrafts(ctx context.Context, days int) (int, error) { return 0, nil }

func (m *MockDraftServiceCancel) GetHistory(ctx context.Context, limit, offset int) ([]repository.BookDraftHistory, int64, error) {
	return nil, 0, nil
}

func TestDraftHandler_CancelDrafts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockDraftServiceCancel)
	handler := NewDraftHandler(mockService)

	mockService.On("CancelDrafts", mock.Anything, []string{"274752", "274781"}).
		Return(2, 3, []error{})

	router := gin.New()
	router.POST("/api/drafts/cancel", handler.CancelDrafts)

	reqBody := map[string]interface{}{
		"ids": []string{"274752", "274781"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/drafts/cancel", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(resp.Body.Bytes(), &response)

	assert.Equal(t, float64(200), response["code"])
	assert.Contains(t, response["message"].(string), "Successfully cancelled")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["cancelled_books"])
	assert.Equal(t, float64(3), data["cancelled_drafts_count"])

	mockService.AssertExpectations(t)
}

func TestDraftHandler_CancelDrafts_EmptyIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockDraftServiceCancel)
	handler := NewDraftHandler(mockService)

	router := gin.New()
	router.POST("/api/drafts/cancel", handler.CancelDrafts)

	reqBody := map[string]interface{}{
		"ids": []string{},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/drafts/cancel", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(resp.Body.Bytes(), &response)

	assert.Equal(t, float64(400), response["code"])
	errObj := response["error"].(map[string]interface{})
	assert.Contains(t, errObj["message"].(string), "ids cannot be empty")
}
