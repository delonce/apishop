package subject

import (
	"github.com/delonce/apishop/internal/service/consumer"
)

//go:generate mockgen -source=subject.go -destination=mocks/mock.go

// Main subject for processing purchase queries
type Subject interface {
	GetSubAmount() int                             // Return amount of current subscribers
	Subscribe(consumer.Consumer)                   // Add new subscriber
	Unsubscribe(consumer.Consumer)                 // Delete some subscriber
	Notify(order map[string]int64) ([]byte, error) // Launch subscribers to process purchase query
}
