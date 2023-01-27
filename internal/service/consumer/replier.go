package consumer

import (
	"context"
	"encoding/json"

	"github.com/delonce/apishop/internal/database"
	"github.com/delonce/apishop/pkg/logging"
)

// Gets json reply to client with his check
type ReplySubscriber struct {
	name      string
	prodDB    database.ProductDB
	logger    *logging.Logger
	reply     chan ClientCheck
	jsonReply chan []byte
}

func GetReplier(name string, prodDB database.ProductDB,
	logger *logging.Logger, clientForm chan ClientCheck, jsChan chan []byte) Consumer {
	return &ReplySubscriber{
		name:      name,
		prodDB:    prodDB,
		logger:    logger,
		reply:     clientForm,
		jsonReply: jsChan,
	}
}

func (rep *ReplySubscriber) GetName() string {
	return rep.name
}

func (rep *ReplySubscriber) Update(ctx context.Context, order map[string]int64) error {
	select {
	case <-ctx.Done():
		// Handle Cancelation
		return ctx.Err()
	case repForm := <-rep.reply:
		if rep.logger != nil {
			rep.logger.Trace("Starting Replier...")
		}

		// Сount all positions
		for name, reqAmount := range order {
			// Get some product from database
			product, err := rep.prodDB.SelectProductByName(name)

			// If error happened writes empty structure in channel
			if err != nil {
				return err
			}

			posCost := product.Cost * reqAmount

			// Add position in list
			repForm.Positions = append(repForm.Positions, ProductPosition{
				Product:   product.Name,
				PosCost:   posCost,
				ReqAmount: reqAmount,
			})

			// Find part
			repForm.TotalSum = repForm.TotalSum + posCost

		}
		// Create json bytes for reply
		return rep.makeJsonReply(repForm)
	}
}

func (rep *ReplySubscriber) makeJsonReply(check ClientCheck) error {
	// Makes json reply for client, structure ClientCheck
	// Send json in []byte channel to subject
	rawBytes, err := json.Marshal(check)

	if err != nil {
		rep.logger.Panicf("Error Marshall, error: %v", err)
		return err
	}

	// Writes in channel
	rep.jsonReply <- rawBytes

	return nil
}
