package produce

import (
	"context"

	"github.com/apfirsov/custom-broker/internal/models"
	"github.com/apfirsov/custom-broker/internal/usecase"
)

type UseCase struct {
	p usecase.Producer
}

func New(p usecase.Producer) *UseCase {
	return &UseCase{
		p: p,
	}
}

func (u *UseCase) Produce(_ context.Context, queueName string, message []byte) error {
	msg := &models.Message{
		Content: message,
	}
	err := u.p.Produce(queueName, msg)

	if err != nil {
		return err
	}

	return nil
}
