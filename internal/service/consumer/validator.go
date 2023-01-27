package consumer

import (
	"context"
	"fmt"

	"github.com/delonce/apishop/internal/database"
	"github.com/delonce/apishop/pkg/logging"
)

// Very important subscriber that validates input data and checks existing nessesary positions in database
// Checks opportunity of making recieved order

type ValidateSubscriber struct {
	name      string
	prodDB    database.ProductDB
	logger    *logging.Logger
	reply     chan ClientCheck
	subAmount int
}

func GetValidateSubscriber(name string, prodDB database.ProductDB,
	logger *logging.Logger, clientForm chan ClientCheck) *ValidateSubscriber {
	return &ValidateSubscriber{
		name:   name,
		prodDB: prodDB,
		logger: logger,
		reply:  clientForm,
	}
}

func (validator *ValidateSubscriber) GetName() string {
	return validator.name
}

func (validator *ValidateSubscriber) Update(ctx context.Context, order map[string]int64) error {
	errString := []string{}
	isConf := true

	for name, reqAmount := range order {
		// Get some product from database
		product, err := validator.prodDB.SelectProductByName(name)

		// If error happened writes empty structure in channel
		if err != nil {
			validator.writeInChan(ClientCheck{})
			return err
		}

		// Creating channel for CheckSender
		// Remembers all order that have more amount of some product than we have
		if product.Amount < reqAmount {
			isConf = false
			errString = append(errString,
				fmt.Sprintf("product: %s, requested_amount: %d, actually amount: %d", name, reqAmount, product.Amount),
			)
		}
	}

	// If all data pass test writes in channel true value
	// validator.isConf <- true
	validator.writeInChan(ClientCheck{
		IsConf: isConf,
		Error:  errString,
	})

	return nil
}

func (validator *ValidateSubscriber) writeInChan(check ClientCheck) {
	for i := 0; i < validator.subAmount; i++ {
		validator.reply <- check
	}
}

func (validator *ValidateSubscriber) SetSubAmount(amount int) {
	// Subcriber amount setter
	validator.subAmount = amount
}
