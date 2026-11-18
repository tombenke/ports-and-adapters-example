package notify

import (
	"context"
	"log/slog"
	"time"

	"github.com/tombenke/ports-and-adapters-example/internal/application/ports"
	"github.com/tombenke/ports-and-adapters-example/internal/domain"
)

// LoggingDecorator adds logging around notification operations.
type LoggingDecorator struct {
	next   ports.NotificationSender
	logger *slog.Logger
}

var _ ports.NotificationSender = (*LoggingDecorator)(nil)

// NewLoggingDecorator wraps a notification sender with structured logging.
func NewLoggingDecorator(next ports.NotificationSender, logger *slog.Logger) *LoggingDecorator {
	return &LoggingDecorator{next: next, logger: logger}
}

// OrderPlaced logs and delegates an order-placed notification.
func (d *LoggingDecorator) OrderPlaced(ctx context.Context, order domain.Order) error {
	start := time.Now()
	err := d.next.OrderPlaced(ctx, order)
	d.logger.Info("notification.OrderPlaced",
		"order_id", order.ID,
		"customer_id", order.CustomerID,
		"duration_ms", time.Since(start).Milliseconds(),
		"error", err,
	)
	return err
}
