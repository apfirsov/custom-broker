package queues_queue_name_messages

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/apfirsov/custom-broker/internal/models"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc producer
}

func New(uc producer) *Handler {
	return &Handler{
		uc: uc,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	queueName := c.Param("queue_name")

	rawData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var jsonData interface{}
	if err = json.Unmarshal(rawData, &jsonData); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid JSON"})
		return
	}

	ctx := c.Request.Context()
	err = h.uc.Produce(ctx, queueName, rawData)

	if errors.Is(err, &models.QueueFullErr{}) {
		c.JSON(http.StatusLocked, gin.H{"error": err.Error()})
		return
	}

	if errors.Is(err, &models.QueueNotFound{}) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
