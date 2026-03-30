package calibre

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/service"
	"github.com/jianyun8023/calibre-api/pkg/response"
)

type DraftHandler struct {
	draftService service.DraftService
}

func NewDraftHandler(draftService service.DraftService) *DraftHandler {
	return &DraftHandler{
		draftService: draftService,
	}
}

func (h *DraftHandler) ReceiveDeletes(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.draftService.ReceiveDeletes(c.Request.Context(), req.IDs); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "delete drafts received successfully", gin.H{"count": len(req.IDs)})
}

func (h *DraftHandler) ReceiveUpdates(c *gin.Context) {
	var req struct {
		Updates []service.BookDraftUpdate `json:"updates" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.draftService.ReceiveUpdates(c.Request.Context(), req.Updates); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "update drafts received successfully", gin.H{"count": len(req.Updates)})
}

func (h *DraftHandler) GetPendingDrafts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		response.BadRequest(c, "Invalid limit parameter")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		response.BadRequest(c, "Invalid offset parameter")
		return
	}

	drafts, total, err := h.draftService.GetPendingDrafts(c.Request.Context(), limit, offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	page := (offset / limit) + 1
	response.Paginated(c, drafts, total, page, limit)
}

func (h *DraftHandler) ApplyDrafts(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	errs := h.draftService.ApplyDrafts(c.Request.Context(), req.IDs)
	if len(errs) > 0 {
		var errorMsgs []string
		for _, err := range errs {
			errorMsgs = append(errorMsgs, err.Error())
		}
		response.SuccessWithMessage(c, "drafts processed with some errors", gin.H{
			"total":  len(req.IDs),
			"errors": errorMsgs,
		})
		return
	}

	response.SuccessWithMessage(c, "drafts applied successfully", gin.H{"count": len(req.IDs)})
}

func (h *DraftHandler) RejectDrafts(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	errs := h.draftService.RejectDrafts(c.Request.Context(), req.IDs)
	if len(errs) > 0 {
		var errorMsgs []string
		for _, err := range errs {
			errorMsgs = append(errorMsgs, err.Error())
		}
		response.SuccessWithMessage(c, "drafts processed with some errors", gin.H{
			"total":  len(req.IDs),
			"errors": errorMsgs,
		})
		return
	}

	response.SuccessWithMessage(c, "drafts rejected successfully", gin.H{"count": len(req.IDs)})
}

func (h *DraftHandler) CancelDrafts(c *gin.Context) {
	var req struct {
		// Note: binding:"required" doesn't reject empty arrays in Gin,
		// so we validate with len() check below
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "ids cannot be empty")
		return
	}

	cancelledBooks, cancelledDrafts, errs := h.draftService.CancelDrafts(c.Request.Context(), req.IDs)

	if len(errs) > 0 {
		var errorMsgs []string
		for _, err := range errs {
			errorMsgs = append(errorMsgs, err.Error())
		}

		if cancelledBooks > 0 {
			response.SuccessWithMessage(c,
				fmt.Sprintf("Cancelled drafts for %d book(s), %d failed", cancelledBooks, len(errs)),
				gin.H{
					"cancelled_books":        cancelledBooks,
					"cancelled_drafts_count": cancelledDrafts,
					"errors":                 errorMsgs,
				})
		} else {
			response.SuccessWithMessage(c,
				"Failed to cancel all drafts",
				gin.H{
					"total":  len(req.IDs),
					"errors": errorMsgs,
				})
		}
		return
	}

	response.SuccessWithMessage(c,
		fmt.Sprintf("Successfully cancelled drafts for %d book(s)", cancelledBooks),
		gin.H{
			"cancelled_books":        cancelledBooks,
			"cancelled_drafts_count": cancelledDrafts,
		})
}

func (h *DraftHandler) GetHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		response.BadRequest(c, "Invalid limit parameter")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		response.BadRequest(c, "Invalid offset parameter")
		return
	}

	histories, total, err := h.draftService.GetHistory(c.Request.Context(), limit, offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	page := (offset / limit) + 1
	response.Paginated(c, histories, total, page, limit)
}
