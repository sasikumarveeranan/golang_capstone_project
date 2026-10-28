package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"order-management/order-service/model"
	"time"

	"github.com/sony/gobreaker"
)

type ProductClient struct {
	baseURL    string
	httpClient *http.Client
	breaker    *gobreaker.CircuitBreaker
}

type productResponse struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
	Available bool    `json:"available"`
}

func NewProductClient(baseURL string) *ProductClient {
	s := gobreaker.Settings{
		Name:        "product-service",
		MaxRequests: 1,
		Interval:    30 * time.Second,
		Timeout:     15 * time.Second,
	}

	return &ProductClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 2 * time.Second},
		breaker:    gobreaker.NewCircuitBreaker(s),
	}
}

// GetProduct gets product details from Product Service.
func (p *ProductClient) GetProduct(ctx context.Context, productID int64) (model.ProductDetails, error) {
	var result productResponse

	v, err := p.breaker.Execute(func() (interface{}, error) {
		var lastErr error
		delay := 100 * time.Millisecond

		for attempt := 0; attempt < 3; attempt++ {
			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodGet,
				fmt.Sprintf("%s/products/%d", p.baseURL, productID),
				nil,
			)
			if err != nil {
				return nil, err
			}

			resp, err := p.httpClient.Do(req)

			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()

				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					return nil, err
				}

				return result, nil
			}

			if resp != nil {
				if resp.StatusCode == http.StatusNotFound {
					resp.Body.Close()
					return nil, errors.New("product not found")
				}

				lastErr = fmt.Errorf(
					"product service returned status %s",
					resp.Status,
				)

				resp.Body.Close()
			} else if err != nil {
				lastErr = err
			}

			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			delay *= 2
		}

		return nil, fmt.Errorf("product service unavailable: %v", lastErr)
	})

	if err != nil {
		return model.ProductDetails{}, err
	}

	result = v.(productResponse)

	return model.ProductDetails{
		ID:    result.ID,
		Name:  result.Name,
		Price: result.Price,
		Stock: result.Stock,
	}, nil
}

// ReduceStock reduces the product stock.
func (p *ProductClient) ReduceStock(
	ctx context.Context,
	productID int64,
	quantity int,
) error {
	return p.changeStock(ctx, productID, quantity, "reduce")
}

// IncreaseStock restores product stock.
// This is used as compensation if order creation fails.
func (p *ProductClient) IncreaseStock(
	ctx context.Context,
	productID int64,
	quantity int,
) error {
	return p.changeStock(ctx, productID, quantity, "increase")
}

func (p *ProductClient) changeStock(
	ctx context.Context,
	productID int64,
	quantity int,
	action string,
) error {

	_, err := p.breaker.Execute(func() (interface{}, error) {

		url := fmt.Sprintf(
			"%s/products/%d/stock/%s?quantity=%d",
			p.baseURL,
			productID,
			action,
			quantity,
		)

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPut,
			url,
			nil,
		)
		if err != nil {
			return nil, err
		}

		resp, err := p.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf(
				"product service returned status %s",
				resp.Status,
			)
		}

		return nil, nil
	})

	return err
}
