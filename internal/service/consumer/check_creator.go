package consumer

import (
	"context"
	"time"

	"github.com/delonce/apishop/internal/database"
	"github.com/delonce/apishop/pkg/logging"
)

type CheckCreatorSubscriber struct {
	name   string
	prodDB database.ProductDB
	logger *logging.Logger
	reply  chan ClientCheck
}

func GetCheckCreator(name string, prodDB database.ProductDB,
	logger *logging.Logger, clientForm chan ClientCheck) Consumer {
	return &CheckCreatorSubscriber{
		name:   name,
		prodDB: prodDB,
		logger: logger,
		reply:  clientForm,
	}
}

func (creator *CheckCreatorSubscriber) GetName() string {
	return creator.name
}

func (creator *CheckCreatorSubscriber) Update(ctx context.Context, order map[string]int64) error {
	select {
	case <-ctx.Done():
		// Handle Cancelation
		return ctx.Err()
	case repForm := <-creator.reply:
		if creator.logger != nil {
			creator.logger.Trace("Starting Check Creator...")
		}

		// Waiting for check from order (PostReader)
		// If empty check end func
		if repForm.IsConf {
			// Initialize new check
			purchCheck := database.Check{
				IsConfirmed: repForm.IsConf,
				DateAt:      time.Now(),
			}

			// Get created check's id
			checkId, err := creator.prodDB.InsertCheck(purchCheck)

			if err != nil {
				return err
			}
			err = creator.addPositions(checkId, order)

			if err != nil {
				return err
			}

			return nil

		} else {
			return nil
		}
	}
}

func (creator *CheckCreatorSubscriber) addPositions(checkId int64, order map[string]int64) error {
	// Add all position from check to order table
	for name, reqAmount := range order {
		// Initialize new position
		order := database.Order{}

		// Get info about some product
		product, err := creator.prodDB.SelectProductByName(name)

		if err != nil {
			return err
		}

		// Fill order fields
		order.CheckID = checkId
		order.ReqAmount = reqAmount
		order.ProductID = product.ID

		// Insert position
		err = creator.prodDB.InsertProductFromCheck(order)

		if err != nil {
			return err
		}
	}

	return nil
}
