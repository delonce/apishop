package app

import (
	"context"

	"github.com/delonce/apishop/internal/config"
	"github.com/delonce/apishop/internal/database"
	pgmanager "github.com/delonce/apishop/internal/database/postgres"
	"github.com/delonce/apishop/internal/delivery"
	"github.com/delonce/apishop/internal/server"
	"github.com/delonce/apishop/internal/service/consumer"
	"github.com/delonce/apishop/internal/service/subject"
	postgresdb "github.com/delonce/apishop/pkg/dbclient"
	"github.com/delonce/apishop/pkg/logging"
	"github.com/julienschmidt/httprouter"
)

type ConsumerApp struct {
	ctx       context.Context
	logger    *logging.Logger
	appConfig *config.Config
}

func NewApp(ctx context.Context, appLog *logging.Logger, cfg *config.Config) *ConsumerApp {
	return &ConsumerApp{
		ctx:       ctx,
		logger:    appLog,
		appConfig: cfg,
	}
}

func (app *ConsumerApp) StartConsumerApplication() {
	prchSubj := app.createPurchaseSubject()
	app.logger.Info("Purchase subject has created")
	router := app.createHTTPRouter(prchSubj)
	app.startHTTPServer(router)
}

func (app *ConsumerApp) startHTTPServer(router *httprouter.Router) {
	app.logger.Info("Getting http server...")
	appServer := server.GetNewServer(app.appConfig.Host, app.appConfig.Port, router)
	app.logger.Infof("Listening on http://%s:%d", app.appConfig.Host, app.appConfig.Port)

	appServer.ListenAndServe()
}

func (app *ConsumerApp) createHTTPRouter(subj subject.Subject) *httprouter.Router {
	transportManager := delivery.NewDeliveryManager(app.logger, subj)

	// Before returning router we need to register urls
	transportManager.Register()

	return transportManager.GetRouter()
}

func (app *ConsumerApp) createPurchaseSubject() subject.Subject {
	// Creating a service that provides customer data

	app.logger.Info("Creating purchase subject")

	// Channel for subscribers (see service/consumer)
	clientReply := make(chan consumer.ClientCheck)
	jsonChannel := make(chan []byte)

	// Get pool of connections for subs
	prodDB := app.initDBPoolConnection()

	// Subject that joins created subscribers
	purchSub := subject.GetPurchaseSubj(app.ctx, app.logger, jsonChannel)

	// Sub creating check of purchase
	sender := consumer.GetCheckCreator("Check Creator", prodDB, app.logger, clientReply)
	// Validating sub
	validator := consumer.GetValidateSubscriber("Validator", prodDB, app.logger, clientReply)
	// Reply client sub
	replier := consumer.GetReplier("Replier", prodDB, app.logger, clientReply, jsonChannel)

	// Process of subscribing
	purchSub.Subscribe(sender)
	purchSub.Subscribe(validator)
	purchSub.Subscribe(replier)

	validator.SetSubAmount(purchSub.GetSubAmount() - 1)

	return purchSub
}

func (app *ConsumerApp) initDBPoolConnection() database.ProductDB {
	// Using pgxpool instead of default sql package
	pgxPool := postgresdb.NewPostgresConnection(app.logger, app.appConfig.DBLogin, app.appConfig.DBPasswd,
		app.appConfig.DBAddr, app.appConfig.DBPort, app.appConfig.DBName)

	storage := pgmanager.NewStorage(app.ctx, pgxPool, app.logger)

	return storage
}
