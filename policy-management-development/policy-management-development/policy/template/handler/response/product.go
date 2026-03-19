package response

import (
	domain "pisapi/core/domain"
	port "pisapi/core/port"
)

type OrderID struct {
	ID int64 `json:"id"`
}

type CreateOrderResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      OrderID `json:"data"`
}

func NewCreateOrderID(id *domain.OrderID) OrderID {
	return OrderID{
		ID: id.ID,
	}
}
