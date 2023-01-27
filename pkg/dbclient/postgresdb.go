package postgresdb

import (
	"context"
	"fmt"

	"github.com/delonce/apishop/pkg/logging"
	"github.com/jackc/pgx/v4/pgxpool"
)

func NewPostgresConnection(logger *logging.Logger, username, password, host, port, database string) *pgxpool.Pool {
	logger.Infof("Connecting to postgresql...")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)

	connPool, err := pgxpool.Connect(context.Background(), connStr)

	if err != nil {
		logger.Panicf("Error opening a connection, %v", err)
	}

	if err = connPool.Ping(context.Background()); err != nil {
		logger.Panicf("Error ping database, %v", err)
	}

	logger.Infof("Connection to postgreSQL is stable")

	return connPool
}
