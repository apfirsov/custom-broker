package subscribe

import (
	"context"

	"github.com/apfirsov/custom-broker/internal/models"
	"github.com/apfirsov/custom-broker/internal/usecase"
)

type UseCase struct {
	s usecase.Subscriber
}

func New(s usecase.Subscriber) *UseCase {
	return &UseCase{
		s: s,
	}
}

func (u *UseCase) Subscribe(ctx context.Context, queueName string) (chan *models.ConsumerMessage, error) {
	chIn, err := u.s.Subscribe(queueName)
	if err != nil {
		return nil, err
	}

	chOut := make(chan *models.ConsumerMessage, 1)

	go func() {
		defer func() {
			err = u.s.Unsubscribe(queueName, chIn)
		}()

		for {
			select {
			case <-ctx.Done():
				close(chOut)
				return
			case msg := <-*chIn:
				chOut <- msg
			}
		}
	}()

	return chOut, nil
}
