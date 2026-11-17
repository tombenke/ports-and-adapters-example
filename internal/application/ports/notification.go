package ports

import (
	"context"

	"github.com/tombenke/ports-and-adapters-example/internal/domain"
)

// NotificationSender sends notifications about order events.
type NotificationSender interface {
	OrderPlaced(ctx context.Context, order domain.Order) error
}
