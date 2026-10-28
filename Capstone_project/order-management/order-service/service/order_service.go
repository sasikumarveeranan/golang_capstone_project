package service

import (
	"context"
	"errors"
	"order-management/order-service/client"
	"order-management/order-service/model"
	"order-management/order-service/repository"
	"time"
)

type OrderService interface {
	Create(context.Context, *model.CreateOrderRequest) (model.Order, error)
	Get(context.Context, int64) (model.Order, error)
	List(context.Context) ([]model.Order, error)
	UpdateStatus(context.Context, int64, model.OrderStatus) error
}

type orderService struct {
	repo     repository.OrderRepository
	products *client.ProductClient
}

func NewOrderService(
	r repository.OrderRepository,
	p *client.ProductClient,
) OrderService {
	return &orderService{
		repo:     r,
		products: p,
	}
}

func (s *orderService) Create(
	ctx context.Context,
	req *model.CreateOrderRequest,
) (model.Order, error) {

	// Step 1: Get product details
	product, err := s.products.GetProduct(ctx, req.ProductID)
	if err != nil {
		return model.Order{}, err
	}

	// Step 2: Check stock
	if product.Stock < req.Quantity {
		return model.Order{}, errors.New("insufficient stock")
	}

	// Step 3: Reduce stock
	if err := s.products.ReduceStock(
		ctx,
		req.ProductID,
		req.Quantity,
	); err != nil {
		return model.Order{}, err
	}

	// Step 4: Create order
	o := model.Order{
		CustomerID:  req.CustomerID,
		ProductID:   req.ProductID,
		ProductName: product.Name,
		Quantity:    req.Quantity,
		Price:       product.Price,
		Status:      model.StatusCreated,
	}

	if err := s.repo.Create(ctx, &o); err != nil {

		// Step 5: Compensation
		// If order creation fails, restore the stock.
		compCtx, cancel := context.WithTimeout(
			context.Background(),
			2*time.Second,
		)
		defer cancel()

		_ = s.products.IncreaseStock(
			compCtx,
			req.ProductID,
			req.Quantity,
		)

		return model.Order{}, err
	}

	return o, nil
}

func (s *orderService) Get(
	ctx context.Context,
	id int64,
) (model.Order, error) {
	return s.repo.Get(ctx, id)
}

func (s *orderService) List(
	ctx context.Context,
) ([]model.Order, error) {
	return s.repo.List(ctx)
}

func (s *orderService) UpdateStatus(
	ctx context.Context,
	id int64,
	status model.OrderStatus,
) error {

	if status != model.StatusCreated &&
		status != model.StatusConfirmed &&
		status != model.StatusCancelled {
		return errors.New("invalid status")
	}

	o, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if o.Status == model.StatusCancelled {
		return errors.New("order is already cancelled")
	}

	return s.repo.UpdateStatus(ctx, id, status)
}
