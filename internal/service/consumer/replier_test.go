package consumer

import (
	"context"
	"errors"
	"testing"

	"github.com/delonce/apishop/internal/database"
	mock_db "github.com/delonce/apishop/internal/database/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestReplierUpdate(t *testing.T) {
	// Test checks Update function in Validator
	type mockBehavior func(s *mock_db.MockProductDB, name string, retProd *database.Product)

	testTable := []struct {
		name          string
		inputOrder    map[string]int64
		product       []*database.Product
		inputCheck    ClientCheck
		expectedJson  []byte
		expectedError error
		mockBehavior  mockBehavior
	}{
		{
			name: "OK",
			inputOrder: map[string]int64{
				"apple": 10,
				"melon": 1,
			},
			product: []*database.Product{
				{
					ID:     1,
					Name:   "apple",
					Cost:   200,
					Amount: 50,
				},

				{
					ID:     2,
					Name:   "melon",
					Cost:   200,
					Amount: 10,
				},
			},
			inputCheck: ClientCheck{
				IsConf: true,
				Error:  []string{},
			},
			expectedJson:  []byte(`{"total_cost":2200,"positions":[{"product":"apple","pos_cost":2000,"req_amount":10},{"product":"melon","pos_cost":200,"req_amount":1}],"is_confirmed":true,"error":[]}`),
			expectedError: nil,
			mockBehavior: func(s *mock_db.MockProductDB, name string, retProd *database.Product) {
				s.EXPECT().SelectProductByName(name).Return(retProd, nil)
			},
		},

		{
			name: "Database Error",
			inputOrder: map[string]int64{
				"apple": 10,
			},
			product: []*database.Product{
				{
					ID:     1,
					Name:   "apple",
					Cost:   200,
					Amount: 50,
				},
			},
			inputCheck: ClientCheck{
				IsConf: true,
				Error:  []string{},
			},
			expectedJson:  []byte(nil),
			expectedError: errors.New("db mock error"),
			mockBehavior: func(s *mock_db.MockProductDB, name string, retProd *database.Product) {
				s.EXPECT().SelectProductByName(name).Return(nil, errors.New("db mock error"))
			},
		},

		{
			name: "Not confirmed check",
			inputOrder: map[string]int64{
				"apple": 10,
				"melon": 1,
			},
			product: []*database.Product{
				{
					ID:     1,
					Name:   "apple",
					Cost:   200,
					Amount: 50,
				},

				{
					ID:     2,
					Name:   "melon",
					Cost:   200,
					Amount: 10,
				},
			},
			inputCheck: ClientCheck{
				IsConf: false,
				Error:  []string{"some mock error 1", "some mock error 2"},
			},
			expectedJson:  []byte(`{"total_cost":2200,"positions":[{"product":"apple","pos_cost":2000,"req_amount":10},{"product":"melon","pos_cost":200,"req_amount":1}],"is_confirmed":false,"error":["some mock error 1","some mock error 2"]}`),
			expectedError: nil,
			mockBehavior: func(s *mock_db.MockProductDB, name string, retProd *database.Product) {
				s.EXPECT().SelectProductByName(name).Return(retProd, nil)
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			prodDB := mock_db.NewMockProductDB(c)

			// Define behavior
			for _, product := range testCase.product {
				testCase.mockBehavior(prodDB, product.Name, product)
			}

			resCheck := make(chan ClientCheck)
			rawBytes := make(chan []byte)

			var resJson []byte

			replier := GetReplier("Test Rep Sub", prodDB, nil, resCheck, rawBytes)

			// Use errgroup to catch errors
			g, ctx := errgroup.WithContext(context.TODO())
			g.Go(func() error {
				// Appeal to Consumer interface, do some subscriber's work
				// Looking for an error
				if err := replier.Update(ctx, testCase.inputOrder); err != nil {
					return err
				}

				return nil
			})

			resCheck <- testCase.inputCheck

			if testCase.expectedError == nil {
				resJson = <-rawBytes
			}

			// Just return error to caller
			err := g.Wait()

			// Assert error
			assert.Equal(t, testCase.expectedJson, resJson)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}
