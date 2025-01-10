package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/apfirsov/custom-broker/internal/api/v1/queues_queue_name_messages"
	"github.com/apfirsov/custom-broker/internal/api/v1/queues_queue_name_subscriptions"
	"github.com/apfirsov/custom-broker/internal/broker"
	"github.com/apfirsov/custom-broker/internal/usecase/produce"
	"github.com/apfirsov/custom-broker/internal/usecase/subscribe"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func main() {
	//  TODO: чтение из конфиг файла
	severPort := "8080"
	queueSize := 3
	queues := []string{
		"app_events",
	}

	logFile, _ := os.Create(fmt.Sprintf("logs_%s.log", time.Now().Format("2006-01-02_15-04-05")))
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	gin.DefaultWriter = multiWriter

	logger := log.New()
	logger.SetOutput(multiWriter)
	logger.Info()

	wsUpgrader := &websocket.Upgrader{}
	wsUpgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	messageBroker := broker.New(logger)
	for _, queueName := range queues {
		err := messageBroker.AddQueue(queueName, queueSize)
		if err != nil {
			logger.Fatal(err)
		}
	}

	produceUC := produce.New(messageBroker)
	subscribeUC := subscribe.New(messageBroker)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.POST("v1/queues/:queue_name/messages", queues_queue_name_messages.New(produceUC).Handle)
	router.GET("v1/queues/:queue_name/subscriptions", queues_queue_name_subscriptions.New(wsUpgrader, subscribeUC).Handle)

	err := router.Run(fmt.Sprintf(":%s", severPort))

	if err != nil {
		logger.Fatal(err)
	}
}
