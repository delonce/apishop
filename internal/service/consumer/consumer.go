package consumer

import "context"

//go:generate mockgen -source=consumer.go -destination=mocks/mock.go

type Consumer interface {
	GetName() string                                          // Return name of subscriber for handy debug and logging
	Update(ctx context.Context, order map[string]int64) error // Do main work
}

// FOR REPLY TO CLIENTS
type ClientCheck struct {
	TotalSum  int64             `json:"total_cost"`
	Positions []ProductPosition `json:"positions"`
	IsConf    bool              `json:"is_confirmed"`
	Error     []string          `json:"error"`
}

type ProductPosition struct {
	Product   string `json:"product"`
	PosCost   int64  `json:"pos_cost"`
	ReqAmount int64  `json:"req_amount"`
}
