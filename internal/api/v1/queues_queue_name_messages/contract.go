package queues_queue_name_messages

import (
	"context"
)

type producer interface {
	Produce(ctx context.Context, queueName string, message []byte) error
}
