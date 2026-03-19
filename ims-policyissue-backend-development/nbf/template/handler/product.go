package handler

import (
	domain "pisapi/core/domain"
	"pisapi/core/port"
	resp "pisapi/handler/response"
	repo "pisapi/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

type ProductHandler struct {
	*serverHandler.Base
	svc *repo.ProductRepository
}

func NewProductHandler(svc *repo.ProductRepository) *ProductHandler {
	base := serverHandler.New("Products").SetPrefix("/v1").AddPrefix("")
	return &ProductHandler{Base: base, svc: svc}
}

func (h *ProductHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/products/order", h.CreateOrder).Name("Create Order"),
	}
}

func (h *ProductHandler) CreateOrder(sctx *serverRoute.Context, req CreateOrderRequest) (*resp.CreateOrderResponse, error) {
	ord := domain.Order{
		UserID:          req.Order.UserID,
		TotalAmount:     req.Order.TotalAmount,
		ShippingAddress: req.Order.ShippingAddress,
		PaymentMethod:   req.Order.PaymentMethod,
	}

	var orderItems []domain.OrderItems
	for _, item := range req.OrderItems {
		orderItems = append(orderItems, domain.OrderItems{
			OrderID:   item.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	product := domain.Product{
		ID:            req.Product.ID,
		StockQuantity: req.Product.StockQuantity,
	}

	u, err := h.svc.CreateOrder(sctx.Ctx, ord, orderItems, product)
	if err != nil {
		log.Error(sctx.Ctx, "Error creating order: %v", err)
		return nil, err
	}
	log.Info(sctx.Ctx, "Order created with ID: %d", u.ID)
	r := &resp.CreateOrderResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data:                 resp.NewCreateOrderID(u),
	}
	return r, nil
}
