package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	apperrors "github.com/jianyun8023/calibre-api/pkg/errors"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestSuccess(t *testing.T) {
	c, w := setupTestContext()

	data := map[string]string{"key": "value"}
	Success(c, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != http.StatusOK {
		t.Errorf("expected code 200, got %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("expected message 'success', got '%s'", resp.Message)
	}
}

func TestError(t *testing.T) {
	c, w := setupTestContext()

	err := apperrors.NewBookNotFoundError("123")
	Error(c, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error info in response")
	}
	if resp.Error.Code != string(apperrors.CodeBookNotFound) {
		t.Errorf("expected error code %s, got %s", apperrors.CodeBookNotFound, resp.Error.Code)
	}
}

func TestPaginated(t *testing.T) {
	c, w := setupTestContext()

	data := []string{"item1", "item2"}
	Paginated(c, data, 100, 1, 10)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Pagination == nil {
		t.Fatal("expected pagination info")
	}
	if resp.Pagination.Total != 100 {
		t.Errorf("expected total 100, got %d", resp.Pagination.Total)
	}
	if resp.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Pagination.Page)
	}
	if resp.Pagination.PageSize != 10 {
		t.Errorf("expected page size 10, got %d", resp.Pagination.PageSize)
	}
	if resp.Pagination.TotalPages != 10 {
		t.Errorf("expected total pages 10, got %d", resp.Pagination.TotalPages)
	}
}

func TestBuilder(t *testing.T) {
	c, w := setupTestContext()

	NewBuilder(c).
		WithCode(http.StatusCreated).
		WithMessage("created").
		WithData(map[string]string{"id": "123"}).
		WithTraceID("trace-123").
		Build()

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Message != "created" {
		t.Errorf("expected message 'created', got '%s'", resp.Message)
	}
	if resp.TraceID != "trace-123" {
		t.Errorf("expected trace_id 'trace-123', got '%s'", resp.TraceID)
	}
}
