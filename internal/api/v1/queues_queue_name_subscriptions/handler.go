package queues_queue_name_subscriptions

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	uc  subscriber
	wsu *websocket.Upgrader
}

func New(wsu *websocket.Upgrader, uc subscriber) *Handler {
	return &Handler{
		uc:  uc,
		wsu: wsu,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	queueName := c.Param("queue_name")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	chMsg, err := h.uc.Subscribe(ctx, queueName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ws, err := h.wsu.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	for msg := range chMsg {
		err = ws.WriteMessage(websocket.TextMessage, msg.GetContent())
		if err != nil {
			break
		}
		msg.Ack()
	}

	c.Status(http.StatusOK)
}
