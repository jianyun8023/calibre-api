package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
)

// Client wraps the Qdrant HTTP API
type Client struct {
	baseURL    string
	collection string
	httpClient *http.Client
}

// NewClient creates a new Qdrant client
func NewClient(url, collection string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    url,
		collection: collection,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Point represents a Qdrant point with vector and payload
type Point struct {
	ID      uint64                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// UpsertPointsRequest represents a batch upsert request
type UpsertPointsRequest struct {
	Points []Point `json:"points"`
}

// UpsertPointsResponse represents the response from upsert
type UpsertPointsResponse struct {
	Result struct {
		OperationID uint64 `json:"operation_id"`
		Status      string `json:"status"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Vector      []float32              `json:"vector"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
	Limit       int                    `json:"limit"`
	WithPayload bool                   `json:"with_payload"`
	WithVector  bool                   `json:"with_vector"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Result []SearchResult `json:"result"`
	Status string         `json:"status"`
	Time   float64        `json:"time"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID      uint64                 `json:"id"`
	Score   float32                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
	Vector  []float32              `json:"vector,omitempty"`
}

// ScrollRequest represents a scroll/scan request
type ScrollRequest struct {
	Limit       int                    `json:"limit"`
	WithPayload bool                   `json:"with_payload"`
	WithVector  bool                   `json:"with_vector"`
	Offset      *uint64                `json:"offset,omitempty"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
}

// ScrollResponse represents scroll results
type ScrollResponse struct {
	Result struct {
		Points         []ScrollPoint `json:"points"`
		NextPageOffset *uint64       `json:"next_page_offset"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

// ScrollPoint represents a point returned from scroll
type ScrollPoint struct {
	ID      uint64                 `json:"id"`
	Payload map[string]interface{} `json:"payload"`
	Vector  []float32              `json:"vector,omitempty"`
}

// CountResponse represents the count response
type CountResponse struct {
	Result struct {
		Count uint64 `json:"count"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

// UpsertPoints inserts or updates points in batch
func (c *Client) UpsertPoints(ctx context.Context, points []Point) error {
	url := fmt.Sprintf("%s/collections/%s/points", c.baseURL, c.collection)

	req := UpsertPointsRequest{Points: points}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upsert failed: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	var upsertResp UpsertPointsResponse
	if err := json.NewDecoder(resp.Body).Decode(&upsertResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if upsertResp.Status != "ok" {
		return fmt.Errorf("upsert status not ok: %s", upsertResp.Status)
	}

	return nil
}

// Search performs vector similarity search
func (c *Client) Search(ctx context.Context, vector []float32, filter map[string]interface{}, limit int) ([]SearchResult, error) {
	url := fmt.Sprintf("%s/collections/%s/points/search", c.baseURL, c.collection)

	req := SearchRequest{
		Vector:      vector,
		Filter:      filter,
		Limit:       limit,
		WithPayload: true,
		WithVector:  false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return searchResp.Result, nil
}

// Scroll retrieves points with pagination
func (c *Client) Scroll(ctx context.Context, limit int, offset *uint64, withVector bool) ([]ScrollPoint, *uint64, error) {
	url := fmt.Sprintf("%s/collections/%s/points/scroll", c.baseURL, c.collection)

	req := ScrollRequest{
		Limit:       limit,
		WithPayload: true,
		WithVector:  withVector,
		Offset:      offset,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("scroll failed: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	var scrollResp ScrollResponse
	if err := json.NewDecoder(resp.Body).Decode(&scrollResp); err != nil {
		return nil, nil, fmt.Errorf("decode response: %w", err)
	}

	return scrollResp.Result.Points, scrollResp.Result.NextPageOffset, nil
}

// Count returns the total number of points in the collection
func (c *Client) Count(ctx context.Context) (uint64, error) {
	url := fmt.Sprintf("%s/collections/%s/points/count", c.baseURL, c.collection)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("count failed: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	var countResp CountResponse
	if err := json.NewDecoder(resp.Body).Decode(&countResp); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return countResp.Result.Count, nil
}

// RetrieveResponse represents the response for a single point retrieval
type RetrieveResponse struct {
	Result struct {
		ID      uint64                 `json:"id"`
		Payload map[string]interface{} `json:"payload"`
		Vector  []float32              `json:"vector"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

// Retrieve gets a single point by ID
func (c *Client) Retrieve(ctx context.Context, id uint64) (*Point, error) {
	url := fmt.Sprintf("%s/collections/%s/points/%d", c.baseURL, c.collection, id)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("retrieve failed: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	var retrieveResp RetrieveResponse
	if err := json.NewDecoder(resp.Body).Decode(&retrieveResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if retrieveResp.Result.ID == 0 && len(retrieveResp.Result.Payload) == 0 {
		return nil, nil
	}

	return &Point{
		ID:      retrieveResp.Result.ID,
		Payload: retrieveResp.Result.Payload,
		Vector:  retrieveResp.Result.Vector,
	}, nil
}

// BookToPayload converts a semantic.Book to Qdrant payload
func BookToPayload(book semantic.Book) map[string]interface{} {
	payload := map[string]interface{}{
		"book_id":       book.ID,
		"title":         book.Title,
		"authors":       book.Authors,
		"author_sort":   book.AuthorSort,
		"publisher":     book.Publisher,
		"isbn":          book.Isbn,
		"rating":        book.Rating,
		"tags":          book.Tags,
		"languages":     book.Languages,
		"comments":      book.Comments,
		"pubdate":       book.PubDate,
		"last_modified": book.LastModified,
		"series_index":  book.SeriesIndex,
		"size":          book.Size,
		"identifiers":   book.Identifiers,
		"cover":         book.Cover,
		"file_path":     book.FilePath,
	}
	return payload
}

// PayloadToBook converts Qdrant payload to semantic.Book
func PayloadToBook(id uint64, payload map[string]interface{}) semantic.Book {
	book := semantic.Book{
		ID: int64(id),
	}

	if v, ok := payload["title"].(string); ok {
		book.Title = v
	}
	if v, ok := payload["author_sort"].(string); ok {
		book.AuthorSort = v
	}
	if v, ok := payload["publisher"].(string); ok {
		book.Publisher = v
	}
	if v, ok := payload["isbn"].(string); ok {
		book.Isbn = v
	}
	if v, ok := payload["rating"].(float64); ok {
		book.Rating = v
	}
	if v, ok := payload["comments"].(string); ok {
		book.Comments = v
	}
	if v, ok := payload["series_index"].(float64); ok {
		book.SeriesIndex = v
	}
	if v, ok := payload["size"].(float64); ok {
		book.Size = int64(v)
	}
	if v, ok := payload["cover"].(string); ok {
		book.Cover = v
	}
	if v, ok := payload["file_path"].(string); ok {
		book.FilePath = v
	}

	// Handle arrays
	if v, ok := payload["authors"].([]interface{}); ok {
		for _, item := range v {
			if str, ok := item.(string); ok {
				book.Authors = append(book.Authors, str)
			}
		}
	}
	if v, ok := payload["tags"].([]interface{}); ok {
		for _, item := range v {
			if str, ok := item.(string); ok {
				book.Tags = append(book.Tags, str)
			}
		}
	}
	if v, ok := payload["languages"].([]interface{}); ok {
		for _, item := range v {
			if str, ok := item.(string); ok {
				book.Languages = append(book.Languages, str)
			}
		}
	}

	// Handle identifiers map
	if v, ok := payload["identifiers"].(map[string]interface{}); ok {
		book.Identifiers = make(map[string]string)
		for k, val := range v {
			if str, ok := val.(string); ok {
				book.Identifiers[k] = str
			}
		}
	}

	// Handle timestamps
	if v, ok := payload["pubdate"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			book.PubDate = t
		}
	}
	if v, ok := payload["last_modified"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			book.LastModified = t
		}
	}

	return book
}
