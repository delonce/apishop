package handlers

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/buger/jsonparser"
	mock_subject "github.com/delonce/apishop/internal/service/subject/mocks"
	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestGetRequest(t *testing.T) {
	testRequestTable := []struct {
		name               string
		inputBody          string
		expectedStatusCode int
		expectedReqBody    string
	}{
		{
			name:               "Default OK",
			inputBody:          "",
			expectedStatusCode: 200,
			expectedReqBody:    `example of POST query: {"order":[{"product":"apple","amount":45},{"product":"melon","amount":11}]}`,
		},

		{
			name:               "OK With Some Body",
			inputBody:          `{"testValue": "testValue"}`,
			expectedStatusCode: 200,
			expectedReqBody:    `example of POST query: {"order":[{"product":"apple","amount":45},{"product":"melon","amount":11}]}`,
		},
	}

	for _, testCase := range testRequestTable {

		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			router := httprouter.New()

			// Empty handler
			transport := &NetworkHandler{
				PurchaseService: nil,
				HandlerLogger:   nil,
				Router:          router,
			}

			// Default path
			router.GET("/", transport.GetHelloPage)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", bytes.NewBufferString(testCase.inputBody))

			// Start server
			router.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedReqBody, w.Body.String())
		})
	}
}

func TestBuyOnePosition(t *testing.T) {
	type mockBehavior func(s *mock_subject.MockSubject, order map[string]int64)

	testRequestTable := []struct {
		name               string
		inputBody          string
		expectedStatusCode int
		expectedReqBody    string
		mockBehavior       mockBehavior
	}{
		{
			name:               "OK JSON",
			inputBody:          `{"order":[{"product":"apple","amount":45},{"product":"melon","amount":11}]}`,
			expectedStatusCode: 200,
			expectedReqBody:    "mock message for success notify",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT().Notify(order).Return([]byte("mock message for success notify"), nil)
			},
		},

		{
			name:               "OK two identical positions in one query",
			inputBody:          `{"order":[{"product":"apple","amount":45},{"product":"apple","amount":45}]}`,
			expectedStatusCode: 200,
			expectedReqBody:    "mock message for success notify",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT().Notify(order).Return([]byte("mock message for success notify"), nil)
			},
		},

		{
			name:               "Not found product id DB",
			inputBody:          `{"order":[{"product":"notexistingproduct","amount":60}]}`,
			expectedStatusCode: 400,
			expectedReqBody:    "{\"critical_error\":\"product with name notexistingproduct doesn't exist\"}",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT().Notify(order).Return([]byte(""), fmt.Errorf("product with name notexistingproduct doesn't exist"))
			},
		},

		{
			name:               "All fields string JSON",
			inputBody:          `{"order":[{"product":"apple","amount":"45"},{"product":"melon","amount":"11"}]}`,
			expectedStatusCode: 400,
			expectedReqBody:    "{\"critical_error\":\"error value in key 'amount', product name: apple\"}{\"critical_error\":\"error value in key 'amount', product name: melon\"}",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT()
			},
		},

		{
			name:               "Empty Body",
			inputBody:          ``,
			expectedStatusCode: 400,
			expectedReqBody:    "{\"critical_error\":\"your order is empty, try to add something in POST query\"}",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT()
			},
		},

		{
			name:               "Empty Body With Key Order",
			inputBody:          `{"order":[]}`,
			expectedStatusCode: 400,
			expectedReqBody:    "{\"critical_error\":\"your order is empty, try to add something in POST query\"}",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT()
			},
		},

		{
			name:               "First wrong Second Right",
			inputBody:          `{"order":[{"name":"apple","amount":45},{"product":"melon","amount":11}]}`,
			expectedStatusCode: 400,
			expectedReqBody:    "{\"critical_error\":\"you need to send string value with key 'product'\"}",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT()
			},
		},

		{
			name:               "Some wrong fields",
			inputBody:          `{"order":[{"x":"1","y":-6,"z":"test"}]`,
			expectedStatusCode: 400,
			expectedReqBody:    "{\"critical_error\":\"you need to send string value with key 'product'\"}",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT()
			},
		},

		{
			name:               "Minus value",
			inputBody:          `{"order":[{"product":"apple","amount":10},{"product":"someMockValue","amount":-5}]}`,
			expectedStatusCode: 400,
			expectedReqBody:    "{\"critical_error\":\"field 'amount' in product someMockValue should be more than 0\"}",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT()
			},
		},

		{
			name:               "Many minus value",
			inputBody:          `{"order":[{"product":"apple","amount":-10},{"product":"someMockValue","amount":-5}]}`,
			expectedStatusCode: 400,
			expectedReqBody:    "{\"critical_error\":\"field 'amount' in product apple should be more than 0\"}{\"critical_error\":\"field 'amount' in product someMockValue should be more than 0\"}",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT()
			},
		},

		{
			name:               "Many wrong fields",
			inputBody:          `{"order":[{"product":"apple","amount":-10},{"product":"someMockValue","amount":"5"},{"errorkey":"melon","amount":"5"}]}`,
			expectedStatusCode: 400,
			expectedReqBody:    "{\"critical_error\":\"field 'amount' in product apple should be more than 0\"}{\"critical_error\":\"error value in key 'amount', product name: someMockValue\"}{\"critical_error\":\"you need to send string value with key 'product'\"}",
			mockBehavior: func(s *mock_subject.MockSubject, order map[string]int64) {
				s.EXPECT()
			},
		},
	}

	for _, testCase := range testRequestTable {

		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			sub := mock_subject.NewMockSubject(c)

			order := map[string]int64{}
			bodyBytes := []byte(testCase.inputBody)

			// Get info from input body and put it in order map
			jsonparser.ArrayEach(bodyBytes, func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
				productName, _ := jsonparser.GetString(value, "product")
				amountProduct, _ := jsonparser.GetInt(value, "amount")

				order[productName] = order[productName] + amountProduct

			}, "order")

			// Starts nessesary behavior
			testCase.mockBehavior(sub, order)

			router := httprouter.New()

			// Handler with PurchaseServer (mock)
			transport := &NetworkHandler{
				PurchaseService: sub,
				HandlerLogger:   nil,
				Router:          router,
			}

			// POST path
			router.POST("/", transport.BuyOnePosition)

			// Send input body
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", bytes.NewBufferString(testCase.inputBody))

			// Start server
			router.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedReqBody, w.Body.String())
		})
	}
}
