package database

import "time"

// table product
type Product struct {
	ID     int64
	Name   string
	Cost   int64
	Amount int64
}

// link table between product and check
type Order struct {
	ID        int64
	CheckID   int64
	ProductID int64
	ReqAmount int64
}

// table check
type Check struct {
	ID           int64
	PurchaseList []Order
	IsConfirmed  bool
	DateAt       time.Time
}
