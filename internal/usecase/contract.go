package usecase

import "github.com/apfirsov/custom-broker/internal/models"

type Producer interface {
	Produce(queueName string, message *models.Message) error
}

type Subscriber interface {
	Subscribe(queueName string) (*chan *models.ConsumerMessage, error)
	Unsubscribe(queueName string, ch *chan *models.ConsumerMessage) error
}
