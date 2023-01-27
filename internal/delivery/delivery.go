package delivery

import (
	"github.com/delonce/apishop/internal/delivery/handlers"
	"github.com/delonce/apishop/internal/service/subject"
	"github.com/delonce/apishop/pkg/logging"

	"github.com/julienschmidt/httprouter"
)

type Delivery interface {
	Register()
	GetRouter() *httprouter.Router
}

type deliveryHandler struct {
	*handlers.NetworkHandler
}

func NewDeliveryManager(logger *logging.Logger, purchService subject.Subject) Delivery {
	return &deliveryHandler{
		&handlers.NetworkHandler{
			PurchaseService: purchService,
			Router:          httprouter.New(),
			HandlerLogger:   logger,
		},
	}
}

func (devHandler *deliveryHandler) Register() {
	// Register all urls of our application
	// Use NetworkHandler struct
	devHandler.HandlerLogger.Info("Starting register handlers")

	devHandler.Router.GET("/", devHandler.GetHelloPage)
	devHandler.Router.POST("/", devHandler.BuyOnePosition)

	devHandler.HandlerLogger.Info("Router had registered all handlers")
}

func (devHandler *deliveryHandler) GetRouter() *httprouter.Router {
	return devHandler.Router
}
