package governance

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListDrafts(c *gin.Context) {
	filter := DraftFilter{
		Limit:  20,
		Offset: 0,
	}

	if status := c.Query("status"); status != "" {
		s := DraftStatus(status)
		filter.Status = &s
	}
	if minStr := c.Query("confidence_min"); minStr != "" {
		if min, err := strconv.ParseFloat(minStr, 64); err == nil {
			filter.ConfidenceMin = &min
		}
	}
	if maxStr := c.Query("confidence_max"); maxStr != "" {
		if max, err := strconv.ParseFloat(maxStr, 64); err == nil {
			filter.ConfidenceMax = &max
		}
	}
	if hasFlags := c.Query("has_flags"); hasFlags == "true" {
		b := true
		filter.HasFlags = &b
	}
	if sessionID := c.Query("session_id"); sessionID != "" {
		filter.SessionID = sessionID
	}
	if bookIDStr := c.Query("book_id"); bookIDStr != "" {
		if bookID, err := strconv.ParseInt(bookIDStr, 10, 64); err == nil {
			filter.BookID = &bookID
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	drafts, total, err := h.service.ListDrafts(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  drafts,
		"total": total,
	})
}

func (h *Handler) GetDraft(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	draft, err := h.service.GetDraft(id)
	if err == ErrDraftNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "draft not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, draft)
}

type reviewRequest struct {
	Version int `json:"version"`
}

func (h *Handler) ApproveDraft(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version required"})
		return
	}

	reviewedBy := c.GetString("user")
	if reviewedBy == "" {
		reviewedBy = "system"
	}

	err = h.service.ApproveDraft(id, req.Version, reviewedBy)
	if err == ErrConcurrentModification {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "concurrent_modification",
			"message": "草稿已被其他用户修改，请刷新后重试",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) RejectDraft(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version required"})
		return
	}

	reviewedBy := c.GetString("user")
	if reviewedBy == "" {
		reviewedBy = "system"
	}

	err = h.service.RejectDraft(id, req.Version, reviewedBy)
	if err == ErrConcurrentModification {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "concurrent_modification",
			"message": "草稿已被其他用户修改，请刷新后重试",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

type updateDraftRequest struct {
	NewValue string `json:"new_value"`
	Version  int    `json:"version"`
}

func (h *Handler) UpdateDraft(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err = h.service.UpdateDraftValue(id, req.NewValue, req.Version)
	if err == ErrConcurrentModification {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "concurrent_modification",
			"message": "草稿已被其他用户修改，请刷新后重试",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

type batchRequest struct {
	Items  []BatchItem `json:"items"`
	Action string      `json:"action"`
}

func (h *Handler) BatchAction(c *gin.Context) {
	var req batchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	reviewedBy := c.GetString("user")
	if reviewedBy == "" {
		reviewedBy = "system"
	}

	var result *BatchResult
	switch req.Action {
	case "approve":
		result = h.service.BatchApprove(req.Items, reviewedBy)
	case "reject":
		result = h.service.BatchReject(req.Items, reviewedBy)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) ApplyDrafts(c *gin.Context) {
	var req ApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	appliedBy := c.GetString("user")
	if appliedBy == "" {
		appliedBy = "system"
	}

	result, err := h.service.ApplyDrafts(req.DraftIDs, appliedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) ApplyAll(c *gin.Context) {
	appliedBy := c.GetString("user")
	if appliedBy == "" {
		appliedBy = "system"
	}

	result, err := h.service.ApplyAllApproved(appliedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListChangelogs(c *gin.Context) {
	filter := ChangelogFilter{
		Limit:  20,
		Offset: 0,
	}

	if bookIDStr := c.Query("book_id"); bookIDStr != "" {
		if bookID, err := strconv.ParseInt(bookIDStr, 10, 64); err == nil {
			filter.BookID = &bookID
		}
	}
	if field := c.Query("field"); field != "" {
		f := MetadataField(field)
		filter.Field = &f
	}
	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse("2006-01-02", fromStr); err == nil {
			filter.FromDate = &from
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse("2006-01-02", toStr); err == nil {
			filter.ToDate = &to
		}
	}
	if reverted := c.Query("reverted"); reverted != "" {
		b := reverted == "true"
		filter.Reverted = &b
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	logs, total, err := h.service.ListChangelogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
	})
}

func (h *Handler) GetChangelog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	log, err := h.service.GetChangelog(id)
	if err == ErrChangelogNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "changelog not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, log)
}

func (h *Handler) RevertChangelog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req RevertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	revertedBy := c.GetString("user")
	if revertedBy == "" {
		revertedBy = "system"
	}

	if err := h.service.RevertChangelog(id, req.Reason, revertedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetSession(c *gin.Context) {
	id := c.Param("id")

	session, err := h.service.GetSession(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}
