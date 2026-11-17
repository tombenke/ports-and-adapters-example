package ports

import (
	"context"

	"github.com/tombenke/ports-and-adapters-example/internal/application/dto"
	"github.com/tombenke/ports-and-adapters-example/internal/domain"
)

// Inbound port interfaces — consumer-defined, one per operation.

type OrderPlacer interface {
	PlaceOrder(ctx context.Context, req dto.PlaceOrderRequest) (domain.Order, error)
}

type OrderGetter interface {
	GetOrder(ctx context.Context, id string) (domain.Order, error)
}

type OrderConfirmer interface {
	ConfirmOrder(ctx context.Context, id string) error
}

type OrderCanceler interface {
	CancelOrder(ctx context.Context, id string) error
}
