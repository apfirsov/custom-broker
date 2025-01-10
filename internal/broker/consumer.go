package broker

import (
	"sync"
	"time"

	"github.com/apfirsov/custom-broker/internal/models"
	"golang.org/x/net/context"
)

type consumer struct {
	mu          sync.Mutex
	notEmpty    *sync.Cond
	subscribers map[*chan *models.ConsumerMessage]bool
	log         logger
}

func newConsumer(l logger) *consumer {
	c := &consumer{
		subscribers: make(map[*chan *models.ConsumerMessage]bool),
		log:         l,
	}
	c.notEmpty = sync.NewCond(&c.mu)
	return c
}

func (c *consumer) subscribe() *chan *models.ConsumerMessage {
	ch := make(chan *models.ConsumerMessage)
	pCh := &ch

	c.mu.Lock()
	defer c.mu.Unlock()

	c.subscribers[pCh] = true
	c.notEmpty.Signal()

	return pCh
}

func (c *consumer) unsubscribe(ch *chan *models.ConsumerMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.subscribers, ch)
}

func (c *consumer) runConsuming(q *queue) {
	for {
		sent := false
		msg := q.get()

		c.mu.Lock()
		for len(c.subscribers) == 0 {
			c.notEmpty.Wait()
		}

		channels := make([]*chan *models.ConsumerMessage, 0, len(c.subscribers))
		for ch := range c.subscribers {
			channels = append(channels, ch)
		}

		c.mu.Unlock()

		consumerMsg := &models.ConsumerMessage{
			Message: msg,
			AckChan: make(chan interface{}, 1),
		}

		for _, ch := range channels {
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer cancel()
				select {
				case *ch <- consumerMsg:
					sent = true
				case <-ctx.Done():
					c.log.Info("канал не готов")
				}
			}()
		}

		if sent {
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				select {
				case <-consumerMsg.AckChan:
					err := q.ack()
					if err != nil {
						c.log.Info("ошибка подтверждения сообщения")
					}
				case <-ctx.Done():
					c.log.Info("таймаут подтверждения, сообщение будет отправлено повторно")
				}
			}()
		}
	}
}
