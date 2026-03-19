package handler

// CreateUserRequest represents the payload to create a user
type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Age       int    `json:"age" validate:"required"`
	City      string `json:"city" validate:"required"`
	Email     string `json:"email" validate:"required"`
}

// UpdateUserRequest represents the payload to update a user (all fields optional)
type UpdateUserRequest struct {
	ID        int64  `uri:"id" validate:"required"`
	FirstName string `json:"first_name" validate:"omitempty"`
	LastName  string `json:"last_name" validate:"omitempty"`
	Age       int    `json:"age" validate:"omitempty"`
	City      string `json:"city" validate:"omitempty"`
	Email     string `json:"email" validate:"omitempty"`
}

// Uri struct for id
type UserIDUri struct {
	ID int64 `uri:"id" validate:"required"`
}

// type ListUsersParams struct {
// 	port.MetadataRequest
// }

// func (p *ListUsersParams) Validate() error {
// 	return nil
// }

type Order struct {
	UserID          int64   `json:"user_id" validate:"required"`
	TotalAmount     float64 `json:"total_amount" validate:"required"`
	ShippingAddress string  `json:"shipping_address" validate:"required"`
	PaymentMethod   string  `json:"payment_method" validate:"required"`
}

type OrderItems struct {
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int64   `json:"quantity"`
	UnitPrice float64 `json:"price"`
}

type Product struct {
	ID            int64 `json:"id"`
	StockQuantity int64 `json:"stock_quantity"`
}

type CreateOrderRequest struct {
	Order      Order        `json:"order" validate:"required"`
	OrderItems []OrderItems `json:"order_items" validate:"required"`
	Product    Product      `json:"product" validate:"required"`
}
