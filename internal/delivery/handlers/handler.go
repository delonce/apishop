package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/buger/jsonparser"
	"github.com/delonce/apishop/internal/service/subject"
	"github.com/delonce/apishop/pkg/logging"
	"github.com/julienschmidt/httprouter"
)

type NetworkHandler struct {
	PurchaseService subject.Subject
	Router          *httprouter.Router
	HandlerLogger   *logging.Logger
}

type JsonErrorReply struct {
	Error string `json:"critical_error"`
}

func (handler *NetworkHandler) GetHelloPage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Example of right POST query
	w.Write([]byte(`example of POST query: {"order":[{"product":"apple","amount":45},{"product":"melon","amount":11}]}`))
}

func (handler *NetworkHandler) BuyOnePosition(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get body of query
	start := time.Now()
	bodyBytes, _ := io.ReadAll(r.Body)

	order, err := createOrderMap(handler.HandlerLogger, w, bodyBytes)

	if err != nil {
		return
	}

	// Catches empty order
	if len(order) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(createJsonErrorReply(handler.HandlerLogger, "your order is empty, try to add something in POST query"))
		return
	}

	// Starts main service (Subject interface, see service/subject)
	reply, err := handler.PurchaseService.Notify(order)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(createJsonErrorReply(handler.HandlerLogger, err.Error()))
		return
	} else {
		w.Write(reply)
		fmt.Println("DURATION ", time.Since(start).Seconds())
		return
	}
}

func createOrderMap(logger *logging.Logger, w http.ResponseWriter, bodyBytes []byte) (map[string]int64, error) {
	// If some error happens while parsing POST query - remembers it and blocks handler
	var validateRequestError error

	// Consider POST query like map with string keys (product) and int64 values (amount)
	order := map[string]int64{}

	// Use jsonparser to parse recieved query
	jsonparser.ArrayEach(bodyBytes, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			validateRequestError = err
			w.WriteHeader(http.StatusBadRequest)
			w.Write(createJsonErrorReply(logger, "you need to send list of products with key 'order'"))
			return
		}

		// Get some name of product
		productName, err := jsonparser.GetString(value, "product")

		if err != nil {
			validateRequestError = err
			w.WriteHeader(http.StatusBadRequest)
			w.Write(createJsonErrorReply(logger, "you need to send string value with key 'product'"))
			return
		}

		// Get nessesary amount of founded product
		amountProduct, err := jsonparser.GetInt(value, "amount")

		if err != nil {
			validateRequestError = err
			w.WriteHeader(http.StatusBadRequest)
			errString := fmt.Sprintf("error value in key 'amount', product name: %s", productName)
			w.Write(createJsonErrorReply(logger, errString))
			return
		}

		// Amount can't be less than 0
		if amountProduct < 0 {
			validateRequestError = errors.New("minus value in field 'amount'")
			w.WriteHeader(http.StatusBadRequest)
			errString := fmt.Sprintf("field 'amount' in product %s should be more than 0", productName)
			w.Write(createJsonErrorReply(logger, errString))
			return
		}

		// Add right product and amount to our map
		// Protect from POST query like [{"product":"apple","amount":2},{"product":"apple","amount":2}]
		order[productName] = order[productName] + amountProduct

	}, "order")

	if validateRequestError != nil {
		return nil, validateRequestError
	}

	return order, nil
}

func createJsonErrorReply(logger *logging.Logger, errorString string) []byte {
	// Just creates json with error if this exists
	jsonRep := JsonErrorReply{Error: errorString}

	// jsonparser doesn't provide func of marshalling json so I prefer to use default encoding/json
	rawBytes, err := json.Marshal(jsonRep)

	if err != nil {
		logger.Panicf("Error Marshall %s, error: %v", errorString, err)
	}

	return rawBytes
}
