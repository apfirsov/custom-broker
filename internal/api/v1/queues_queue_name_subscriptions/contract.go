package queues_queue_name_subscriptions

import (
	"context"

	"github.com/apfirsov/custom-broker/internal/models"
)

type subscriber interface {
	Subscribe(ctx context.Context, queueName string) (chan *models.ConsumerMessage, error)
}
