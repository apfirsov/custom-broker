package broker

import (
	"errors"
	"sync"

	"github.com/apfirsov/custom-broker/internal/models"
)

type queue struct {
	mu         sync.Mutex
	notEmpty   *sync.Cond
	elements   []*models.Message
	size       int
	head, tail int
	count      int
	log        logger
}

func (q *queue) push(item *models.Message) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.count == q.size {
		return &models.QueueFullErr{}
	}

	q.elements[q.tail] = item
	q.tail = (q.tail + 1) % q.size
	q.count++

	q.notEmpty.Signal()

	return nil
}

func (q *queue) get() *models.Message {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.count == 0 {
		q.notEmpty.Wait()
	}

	return q.elements[q.head]
}

func (q *queue) ack() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.count == 0 {
		return errors.New("подтверждать нечего")
	}

	q.head = (q.head + 1) % q.size
	q.count--

	return nil
}

func newQueue(size int, log logger) *queue {
	q := &queue{
		elements: make([]*models.Message, size),
		size:     size,
		head:     0,
		tail:     0,
		count:    0,
		log:      log,
	}
	q.notEmpty = sync.NewCond(&q.mu)

	return q
}
