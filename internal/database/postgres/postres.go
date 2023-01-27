package pgmanager

import (
	"context"
	"fmt"

	"github.com/delonce/apishop/internal/database"
	"github.com/delonce/apishop/pkg/logging"
	"github.com/jackc/pgx/v4/pgxpool"
)

type postgresDB struct {
	ctx       context.Context
	dbmanager *pgxpool.Pool
	logger    *logging.Logger
}

func NewStorage(ctx context.Context, pool *pgxpool.Pool, logger *logging.Logger) database.ProductDB {
	return &postgresDB{
		ctx:       ctx,
		dbmanager: pool,
		logger:    logger,
	}
}

func (pgdb *postgresDB) SelectProductByName(productName string) (*database.Product, error) {
	queryString := `
		SELECT id, name, cost, amount FROM product WHERE name=$1
	`

	// Trace every query in logs to handy processing
	pgdb.logger.Trace("SQL Query: ", queryString)

	// Get necessary model
	product := database.Product{}

	err := pgdb.dbmanager.QueryRow(pgdb.ctx, queryString, productName).Scan(&product.ID, &product.Name, &product.Cost, &product.Amount)

	if err != nil {
		// Process DB errors in this part of code
		// Because of using goroutines in service we cannot handle error and describe it there as well as we do here
		// Attemts of catch most common errors
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("product with name %s doesn't exist", productName)
		} else {
			pgdb.logger.Errorf("error when trying buy %s, error: %v", productName, err)
			return nil, err
		}
	}

	return &product, nil
}

func (pgdb *postgresDB) InsertCheck(purchCheck database.Check) (int64, error) {
	// Just insert structure Check into table "check"
	// Generates errors if it occurs
	queryString := `
		INSERT INTO "check"
			(is_confirmed, date)
		VALUES 
			($1, $2)
		RETURNING id
	`

	pgdb.logger.Trace("SQL Query: ", queryString)
	err := pgdb.dbmanager.QueryRow(pgdb.ctx, queryString, purchCheck.IsConfirmed, purchCheck.DateAt).Scan(&purchCheck.ID)

	if err != nil {
		pgdb.logger.Errorf("error when trying insert check, error: %v", err)
		return 0, err
	}

	return purchCheck.ID, nil
}

func (pgdb *postgresDB) InsertProductFromCheck(position database.Order) error {
	// Just insert structure Check into table "order"
	// Generates errors if it occurs
	queryString := `
		INSERT INTO "order"
			(product_id, check_id, req_amount)
		VALUES 
			($1, $2, $3)
		RETURNING id
	`

	pgdb.logger.Trace("SQL Query: ", queryString)
	_, err := pgdb.dbmanager.Exec(pgdb.ctx, queryString, position.ProductID, position.CheckID, position.ReqAmount)

	if err != nil {
		pgdb.logger.Errorf("error when trying insert position, error: %v", err)
		return err
	}

	return nil
}
