package broker

import (
	"fmt"

	"github.com/apfirsov/custom-broker/internal/models"
)

type entity struct {
	queue    *queue
	consumer *consumer
}

type Broker struct {
	entities map[string]*entity
	log      logger
}

func New(log logger) *Broker {
	return &Broker{
		entities: make(map[string]*entity),
		log:      log,
	}
}

func (b *Broker) AddQueue(name string, size int) error {
	if _, has := b.entities[name]; has {
		return fmt.Errorf("очередь %s уже существует", name)
	}

	e := &entity{
		queue:    newQueue(size, b.log),
		consumer: newConsumer(b.log),
	}
	b.entities[name] = e

	go e.consumer.runConsuming(e.queue)

	return nil
}

func (b *Broker) Produce(queueName string, message *models.Message) error {
	e, has := b.entities[queueName]
	if !has {
		return &models.QueueNotFound{}
	}

	err := e.queue.push(message)
	if err != nil {
		return err
	}

	return nil
}

func (b *Broker) Subscribe(queueName string) (*chan *models.ConsumerMessage, error) {
	e, has := b.entities[queueName]
	if !has {
		return nil, fmt.Errorf("queue not found")
	}

	return e.consumer.subscribe(), nil
}

func (b *Broker) Unsubscribe(queueName string, ch *chan *models.ConsumerMessage) error {
	e, has := b.entities[queueName]
	if !has {
		return fmt.Errorf("queue not found")
	}

	e.consumer.unsubscribe(ch)

	return nil
}
