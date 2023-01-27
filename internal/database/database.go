package database

// Interface provides interaction with database
//go:generate mockgen -source=database.go -destination=mocks/mock.go
type ProductDB interface {
	SelectProductByName(productName string) (*Product, error)

	InsertCheck(purchCheck Check) (int64, error)
	InsertProductFromCheck(position Order) error
}
