package domain

type Order struct {
	UserID          int64   `db:"user_id"`
	TotalAmount     float64 `db:"total_amount"`
	ShippingAddress string  `db:"shipping_address"`
	PaymentMethod   string  `db:"payment_method"`
}

type OrderItems struct {
	OrderID   int64   `db:"order_id"`
	ProductID int64   `db:"product_id"`
	Quantity  int64   `db:"quantity"`
	UnitPrice float64 `db:"unit_price"`
}

type Product struct {
	ID            int64 `db:"id"`
	StockQuantity int64 `db:"stock_quantity"`
}

type OrderID struct {
	ID int64 `db:"id"`
}
