package handler

import (
	"net/http"

	repository "github.com/bimal009/Zovly/internal/repo"
	"github.com/bimal009/Zovly/pkg/responses"
	"github.com/gin-gonic/gin"
)

type InboxHandler struct {
	conversationRepo repository.ConversationRepo
	messageRepo      repository.MessageRepo
}

func NewInboxHandler(conversationRepo repository.ConversationRepo, messageRepo repository.MessageRepo) *InboxHandler {
	return &InboxHandler{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
	}
}

func (h *InboxHandler) ListConversations(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	limit, offset := paginationFromQuery(c)
	convs, total, err := h.conversationRepo.ListByBusiness(c.Request.Context(), businessID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to list conversations"))
		return
	}

	c.JSON(http.StatusOK, responses.Paginated("conversations fetched", convs, total, limit, offset))
}

func (h *InboxHandler) GetMessages(c *gin.Context) {
	businessID, ok := businessIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, responses.Unauthorized("unauthorized"))
		return
	}

	conversationID := c.Param("id")

	if _, err := h.conversationRepo.GetByID(c.Request.Context(), conversationID, businessID); err != nil {
		c.JSON(http.StatusNotFound, responses.NotFound("conversation not found"))
		return
	}

	limit, _ := paginationFromQuery(c)
	msgs, err := h.messageRepo.GetByConversation(c.Request.Context(), conversationID, nil, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.InternalServerError("failed to fetch messages"))
		return
	}

	c.JSON(http.StatusOK, responses.Success("messages fetched", msgs))
}
