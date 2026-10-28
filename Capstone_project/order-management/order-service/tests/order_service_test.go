package tests

import (
	"context"
	"order-management/order-service/model"
	"order-management/order-service/repository"
	"testing"
)

type fakeRepo struct{}

func (f *fakeRepo) Create(context.Context, *model.Order) error {
	return nil
}
func (f *fakeRepo) Get(context.Context, int64) (model.Order, error) {
	return model.Order{ID: 1}, nil
}
func (f *fakeRepo) List(context.Context) ([]model.Order, error) {
	return []model.Order{}, nil
}
func (f *fakeRepo) UpdateStatus(context.Context, int64, model.OrderStatus) error {
	return nil
}

var _ repository.OrderRepository = &fakeRepo{}

func TestStatusValues(t *testing.T) {
	if model.StatusCreated != "CREATED" || model.StatusConfirmed != "CONFIRMED" || model.StatusCancelled != "CANCELLED" {
		t.Fatal("invalid status constants")
	}
}
