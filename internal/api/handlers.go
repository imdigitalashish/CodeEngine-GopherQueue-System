package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imdigitalashish/QueueSystemsInGolang/internal/queue"
)

type Handler struct {
	Queue *queue.Queue
}

func NewHandler(q *queue.Queue) *Handler {
	return &Handler{
		Queue: q,
	}
}

func (h *Handler) ProcessContent(c *gin.Context) {
	content := c.PostForm("content")
	language := c.PostForm("language")
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
		return
	}

	id := h.Queue.AddJob(language, content)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handler) CheckStatus(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	job, exists := h.Queue.CheckStatus(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}
