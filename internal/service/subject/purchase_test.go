package subject

import (
	"context"
	"errors"
	"fmt"
	"testing"

	mock_consumer "github.com/delonce/apishop/internal/service/consumer/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestEmptySubscribers(t *testing.T) {
	// Test checks behavior of subject if it doesn't have any subscribers
	testTable := []struct {
		name          string
		inputOrder    map[string]int64
		expectedReply []byte
		expectedError error
	}{
		{
			name: "Empty Subs",
			inputOrder: map[string]int64{
				"apple": 10,
				"milk":  14,
			},
			expectedError: errors.New("subject doen't have any subscribers"),
			expectedReply: nil,
		},
	}

	for _, testCase := range testTable {

		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			// Init net context
			ctx := context.Background()

			// Init channel between subject and mock consumer
			jsChan := make(chan []byte)

			// Init new subject and subcribe mock consumer
			subject := GetPurchaseSubj(ctx, nil, jsChan)

			rawBytes, err := subject.Notify(testCase.inputOrder)

			// Asserts
			assert.Equal(t, testCase.expectedError, err)
			assert.Equal(t, testCase.expectedReply, rawBytes)
		})
	}
}

func TestNotify(t *testing.T) {
	type updateBehavior func(s *mock_consumer.MockConsumer, ctx context.Context, order map[string]int64)
	type nameBehavior func(s *mock_consumer.MockConsumer)

	testTable := []struct {
		name                 string
		inputOrder           map[string]int64
		expectedReply        []byte
		expectedError        error
		internalConsumerJson []byte
		updateBehavior       updateBehavior
		nameBehavior         nameBehavior
	}{
		{
			name: "OK",
			inputOrder: map[string]int64{
				"apple": 10,
				"milk":  14,
			},
			internalConsumerJson: []byte("mock message for success purchase service"),
			expectedError:        nil,
			expectedReply:        []byte("mock message for success purchase service"),
			updateBehavior: func(s *mock_consumer.MockConsumer, ctx context.Context, order map[string]int64) {
				s.EXPECT().Update(ctx, order).Return(nil)
			},
			nameBehavior: func(s *mock_consumer.MockConsumer) {
				s.EXPECT().GetName().Return("Mock Consumer")
			},
		},

		{
			name: "Some mock error",
			inputOrder: map[string]int64{
				"apple": 10,
				"milk":  14,
			},
			internalConsumerJson: []byte("mock message for success purchase service"),
			expectedError:        errors.New("some mock error"),
			expectedReply:        nil,
			updateBehavior: func(s *mock_consumer.MockConsumer, ctx context.Context, order map[string]int64) {
				s.EXPECT().Update(ctx, order).Return(errors.New("some mock error"))
			},
			nameBehavior: func(s *mock_consumer.MockConsumer) {
				s.EXPECT().GetName().Return("Mock Consumer")
			},
		},
	}

	for _, testCase := range testTable {

		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			// Init new Consumer
			con := mock_consumer.NewMockConsumer(c)
			// Init net context
			ctx := context.Background()

			// Create context for Consumer's Update function
			_, updctx := errgroup.WithContext(ctx)

			// Init channel between subject and mock consumer
			jsChan := make(chan []byte)

			// Start nessesary behavior
			testCase.nameBehavior(con)
			testCase.updateBehavior(con, updctx, testCase.inputOrder)

			// Init new subject and subcribe mock consumer
			subject := GetPurchaseSubj(ctx, nil, jsChan)
			subject.Subscribe(con)

			// Writes in channel for subject, expected that subject return same information that was sended in channel
			// If error in Update function was nil
			go func() {
				jsChan <- testCase.internalConsumerJson
			}()

			rawBytes, err := subject.Notify(testCase.inputOrder)

			// Asserts
			assert.Equal(t, testCase.expectedError, err)
			assert.Equal(t, testCase.expectedReply, rawBytes)
		})
	}
}

func TestSomeSubs(t *testing.T) {
	type nameBehavior func(s *mock_consumer.MockConsumer, id int)

	testTable := []struct {
		name          string
		expectedReply int
		nameBehavior  nameBehavior
	}{
		{
			name:          "OK",
			expectedReply: 3,
			nameBehavior: func(s *mock_consumer.MockConsumer, id int) {
				s.EXPECT().GetName().Return(fmt.Sprintf("id: %d", id))
			},
		},

		{
			name:          "OK",
			expectedReply: 9,
			nameBehavior: func(s *mock_consumer.MockConsumer, id int) {
				s.EXPECT().GetName().Return(fmt.Sprintf("id: %d", id))
			},
		},

		{
			name:          "OK",
			expectedReply: 30,
			nameBehavior: func(s *mock_consumer.MockConsumer, id int) {
				s.EXPECT().GetName().Return(fmt.Sprintf("id: %d", id))
			},
		},
	}

	for _, testCase := range testTable {

		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			// Init net context
			ctx := context.Background()

			// Init new subject and subcribe mock consumer
			subject := GetPurchaseSubj(ctx, nil, make(chan []byte))

			for i := 0; i < testCase.expectedReply; i++ {
				// Init new Consumer
				con := mock_consumer.NewMockConsumer(c)

				// Start nessesary behavior
				testCase.nameBehavior(con, i)

				subject.Subscribe(con)
			}

			subAmount := subject.GetSubAmount()

			// Asserts
			assert.Equal(t, testCase.expectedReply, subAmount)
		})
	}
}
