package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"order-management/order-service/model"
)

type OrderRepository interface {
	Create(context.Context, *model.Order) error
	Get(context.Context, int64) (model.Order, error)
	List(context.Context) ([]model.Order, error)
	UpdateStatus(context.Context, int64, model.OrderStatus) error
}

type orderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(
	ctx context.Context,
	o *model.Order,
) error {

	return r.db.QueryRow(
		ctx,
		`INSERT INTO orders
			(customer_id, product_id, product_name, quantity, price, status)
		 VALUES
			($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		o.CustomerID,
		o.ProductID,
		o.ProductName,
		o.Quantity,
		o.Price,
		o.Status,
	).Scan(
		&o.ID,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
}

func (r *orderRepository) Get(
	ctx context.Context,
	id int64,
) (model.Order, error) {

	var o model.Order

	err := r.db.QueryRow(
		ctx,
		`SELECT
			id,
			customer_id,
			product_id,
			product_name,
			quantity,
			price::float8,
			status,
			created_at,
			updated_at
		 FROM orders
		 WHERE id = $1`,
		id,
	).Scan(
		&o.ID,
		&o.CustomerID,
		&o.ProductID,
		&o.ProductName,
		&o.Quantity,
		&o.Price,
		&o.Status,
		&o.CreatedAt,
		&o.UpdatedAt,
	)

	return o, err
}

func (r *orderRepository) List(
	ctx context.Context,
) ([]model.Order, error) {

	rows, err := r.db.Query(
		ctx,
		`SELECT
			id,
			customer_id,
			product_id,
			product_name,
			quantity,
			price::float8,
			status,
			created_at,
			updated_at
		 FROM orders
		 ORDER BY created_at DESC`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	out := []model.Order{}

	for rows.Next() {
		var o model.Order

		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.ProductID,
			&o.ProductName,
			&o.Quantity,
			&o.Price,
			&o.Status,
			&o.CreatedAt,
			&o.UpdatedAt,
		); err != nil {
			return nil, err
		}

		out = append(out, o)
	}

	return out, rows.Err()
}

func (r *orderRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	status model.OrderStatus,
) error {

	tag, err := r.db.Exec(
		ctx,
		`UPDATE orders
		 SET status = $1,
		     updated_at = NOW()
		 WHERE id = $2`,
		status,
		id,
	)

	if err == nil && tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return err
}
