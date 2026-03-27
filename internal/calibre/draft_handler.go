package calibre

import (
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
	drafts, err := h.draftService.GetPendingDrafts(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, drafts)
}

func (h *DraftHandler) ApplyDrafts(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.draftService.ApplyDrafts(c.Request.Context(), req.IDs); err != nil {
		response.Error(c, err)
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

	if err := h.draftService.RejectDrafts(c.Request.Context(), req.IDs); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "drafts rejected successfully", gin.H{"count": len(req.IDs)})
}
