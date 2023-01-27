package subject

import (
	"context"
	"errors"

	"github.com/delonce/apishop/internal/service/consumer"
	"github.com/delonce/apishop/pkg/logging"
	"golang.org/x/sync/errgroup"
)

type PurchaseSubject struct {
	ctx             context.Context
	logger          *logging.Logger
	consumers       map[string]consumer.Consumer
	jsonClientReply chan []byte
}

func GetPurchaseSubj(ctx context.Context, logger *logging.Logger, jsChan chan []byte) Subject {
	return &PurchaseSubject{
		ctx:             ctx,
		logger:          logger,
		consumers:       make(map[string]consumer.Consumer),
		jsonClientReply: jsChan,
	}
}

func (subject *PurchaseSubject) Subscribe(subscriber consumer.Consumer) {
	subject.consumers[subscriber.GetName()] = subscriber
}

func (subject *PurchaseSubject) Unsubscribe(subscriber consumer.Consumer) {
	delete(subject.consumers, subscriber.GetName())
}

func (subject *PurchaseSubject) Notify(order map[string]int64) ([]byte, error) {
	if len(subject.consumers) == 0 {
		return nil, errors.New("subject doen't have any subscribers")
	}
	// Use errgroup to catch errors
	g, ctx := errgroup.WithContext(subject.ctx)
	for _, sub := range subject.consumers {
		callFunc := sub
		// Concurrent launch some subscriber
		g.Go(func() error {
			// Appeal to Consumer interface, do some subscriber's work
			// Looking for an error
			if err := callFunc.Update(ctx, order); err != nil {
				return err
			}

			return nil
		})
	}

	// Waiting for creating json reply to client
	jsonBytes := <-subject.jsonClientReply

	// Just return error to caller
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return jsonBytes, nil
}

func (subject *PurchaseSubject) GetSubAmount() int {
	return len(subject.consumers)
}
