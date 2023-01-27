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

func TestValidatorUpdate(t *testing.T) {
	// Test checks Update function in Validator
	type mockBehavior func(s *mock_db.MockProductDB, name string, retProd *database.Product)

	testTable := []struct {
		name          string
		inputOrder    map[string]int64
		product       []*database.Product
		expectedCheck ClientCheck
		expectedError error
		mockBehavior  mockBehavior
	}{
		{
			name: "OK",
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
			expectedCheck: ClientCheck{
				IsConf: true,
				Error:  []string{},
			},
			expectedError: nil,
			mockBehavior: func(s *mock_db.MockProductDB, name string, retProd *database.Product) {
				s.EXPECT().SelectProductByName(name).Return(retProd, nil)
			},
		},

		{
			name: "Too big amount",
			inputOrder: map[string]int64{
				"apple": 200,
			},
			product: []*database.Product{
				{
					ID:     1,
					Name:   "apple",
					Cost:   200,
					Amount: 50,
				},
			},
			expectedCheck: ClientCheck{
				IsConf: false,
				Error:  []string{"product: apple, requested_amount: 200, actually amount: 50"},
			},
			expectedError: nil,
			mockBehavior: func(s *mock_db.MockProductDB, name string, retProd *database.Product) {
				s.EXPECT().SelectProductByName(name).Return(retProd, nil)
			},
		},

		{
			name: "Error from database",
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
			expectedCheck: ClientCheck{},
			expectedError: errors.New("db mock error"),
			mockBehavior: func(s *mock_db.MockProductDB, name string, retProd *database.Product) {
				s.EXPECT().SelectProductByName(name).Return(nil, errors.New("db mock error"))
			},
		},

		{
			name: "Few products in order",
			inputOrder: map[string]int64{
				"apple": 10,
				"melon": 11,
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
			expectedCheck: ClientCheck{
				IsConf: false,
				Error:  []string{"product: melon, requested_amount: 11, actually amount: 10"},
			},
			expectedError: nil,
			mockBehavior: func(s *mock_db.MockProductDB, name string, retProd *database.Product) {
				s.EXPECT().SelectProductByName(name).Return(retProd, nil)
			},
		},

		{
			name: "Few products in order №2",
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
			expectedCheck: ClientCheck{
				IsConf: true,
				Error:  []string{},
			},
			expectedError: nil,
			mockBehavior: func(s *mock_db.MockProductDB, name string, retProd *database.Product) {
				s.EXPECT().SelectProductByName(name).Return(retProd, nil)
			},
		},

		{
			name: "Few products in order №2",
			inputOrder: map[string]int64{
				"apple": 200,
				"melon": 200,
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
			expectedCheck: ClientCheck{
				IsConf: false,
				Error:  []string{"product: apple, requested_amount: 200, actually amount: 50", "product: melon, requested_amount: 200, actually amount: 10"},
			},
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

			validator := GetValidateSubscriber("Test Val Sub", prodDB, nil, resCheck)
			validator.SetSubAmount(1)

			// Use errgroup to catch errors
			g, ctx := errgroup.WithContext(context.TODO())
			g.Go(func() error {
				// Appeal to Consumer interface, do some subscriber's work
				// Looking for an error
				if err := validator.Update(ctx, testCase.inputOrder); err != nil {
					return err
				}

				return nil
			})

			// Waiting for creating json reply to client
			result := <-resCheck

			// Just return error to caller
			err := g.Wait()

			// Assert error
			assert.Equal(t, testCase.expectedCheck.IsConf, result.IsConf)
			assert.Equal(t, testCase.expectedCheck.Error, result.Error)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}
