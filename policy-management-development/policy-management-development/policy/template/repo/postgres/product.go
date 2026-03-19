package repo

import (
	"context"

	"pisapi/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

type ProductRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

func NewProductRepository(db *dblib.DB, cfg *config.Config) *ProductRepository {
	return &ProductRepository{db: db, cfg: cfg}
}

func (r *ProductRepository) CreateOrder(ctx context.Context, order domain.Order, orderItems []domain.OrderItems, product domain.Product) (*domain.OrderID, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	batch := &pgx.Batch{}

	query1 := dblib.Psql.Insert("orders").
		Columns("user_id", "total_amount", "shipping_address", "payment_method").
		Values(order.UserID, order.TotalAmount, order.ShippingAddress, order.PaymentMethod).
		Suffix("RETURNING id")

	var id domain.OrderID

	dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.OrderID], &id)

	query2 := dblib.Psql.Insert("order_items").
		Columns("order_id", "product_id", "quantity", "unit_price")

	for _, item := range orderItems {
		query2 = query2.Values(item.OrderID, item.ProductID, item.Quantity, item.UnitPrice)
	}

	dblib.QueueExecRow(batch, query2)

	query3 := dblib.Psql.Update("products").
		Set("stock_quantity", sq.Expr("stock_quantity - ?", product.StockQuantity)).
		Where(sq.Eq{"id": product.ID})

	dblib.QueueExecRow(batch, query3)

	err := r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	return &id, nil

}
