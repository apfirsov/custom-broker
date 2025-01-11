package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	produceUrlTemplate = "http://localhost:8080/v1/queues/%v/messages"
	consumeUrlTemplate = "ws://localhost:8080/v1/queues/%v/subscriptions"
)

type message struct {
	ID       string        `json:"id"`
	Queue    string        `json:"queue"`
	TimeIn   time.Time     `json:"time_in"`
	Duration time.Duration `json:"duration"`
}

func saveMessage(storage map[string]*message, ch chan *message, wg *sync.WaitGroup) {
	defer wg.Done()
	for msg := range ch {
		storage[msg.ID] = msg
	}
}

func transferMsg(wg *sync.WaitGroup, in chan *message, out chan *message) {
	defer wg.Done()
	for msg := range in {
		out <- msg
	}
}

func sendMessage(url string, message *message) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("ошибка при конвертации в JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	//fmt.Println(resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("статус ответа %v, %v", resp.Status, body)
	}

	return nil
}

func producer(ctx context.Context, queueName string, ch chan *message) {
	defer close(ch)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(100 * time.Millisecond)
			msg := &message{
				ID:     uuid.New().String(),
				Queue:  queueName,
				TimeIn: time.Now(),
			}
			urlStr := fmt.Sprintf(produceUrlTemplate, queueName)

			err := sendMessage(urlStr, msg)
			if err != nil {
				//log.Error(err)
				continue
			}
			ch <- msg
		}
	}
}

func consumer(ctx context.Context, queueName string, ch chan *message) {
	defer close(ch)

	urlStr := fmt.Sprintf(consumeUrlTemplate, queueName)
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Error(err)
		return
	}

	// Подключение к WebSocket серверу
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Чтение сообщений из WebSocket
			_, msgIn, err := conn.ReadMessage()
			if err != nil {
				log.Error(err)
				return
			}

			msg := &message{}
			err = json.Unmarshal(msgIn, msg)
			if err != nil {
				log.Error(err)
				continue
			}
			msg.Duration = time.Since(msg.TimeIn)
			ch <- msg
		}
	}
}

func main() {
	queueCount := 3
	producerCount := 1000
	consumerCount := 10
	testDuration := 60

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sentMsg := make(map[string]*message)
	receivedMsg := make(map[string]*message)
	sentCh := make(chan *message, 10^8)
	receivedCh := make(chan *message, 10^8)

	transferWG := &sync.WaitGroup{}
	for i := 1; i < queueCount+1; i++ {
		queueName := fmt.Sprintf("app_events_%v", i)

		for i := 0; i < producerCount; i++ {
			producerCh := make(chan *message, 10^3)
			transferWG.Add(1)
			go transferMsg(transferWG, producerCh, sentCh)
			go producer(ctx, queueName, producerCh)
		}

		for i := 0; i < consumerCount; i++ {
			consumerCh := make(chan *message, 10^3)
			transferWG.Add(1)
			go transferMsg(transferWG, consumerCh, receivedCh)
			go consumer(ctx, queueName, consumerCh)
		}
	}

	saveWG := &sync.WaitGroup{}
	saveWG.Add(2)
	go saveMessage(sentMsg, sentCh, saveWG)
	go saveMessage(receivedMsg, receivedCh, saveWG)

	timer := time.After(time.Second * time.Duration(testDuration))
	<-timer
	cancel()
	transferWG.Wait()
	close(sentCh)
	close(receivedCh)
	saveWG.Wait()

	fmt.Println(len(sentMsg))
	fmt.Println(len(receivedMsg))
	durSum := time.Duration(0)
	if len(receivedMsg) > 0 {
		for k := range receivedMsg {
			durSum += receivedMsg[k].Duration
		}
		avg := durSum / time.Duration(len(receivedMsg))
		fmt.Println(avg.Microseconds())
	}
}
