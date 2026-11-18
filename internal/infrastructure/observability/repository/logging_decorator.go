package repository

import (
	"context"
	"time"

	"log/slog"

	"github.com/tombenke/ports-and-adapters-example/internal/application/ports"
	"github.com/tombenke/ports-and-adapters-example/internal/domain"
)

// LoggingDecorator is a repository decorator that adds logging around the repository operations.
type LoggingDecorator struct {
	next   ports.OrderRepository
	logger *slog.Logger
}

// Ensure LoggingDecorator implements the OrderRepository interface.
var _ ports.OrderRepository = (*LoggingDecorator)(nil)

// NewLoggingDecorator creates a new instance of LoggingDecorator.
func NewLoggingDecorator(next ports.OrderRepository, logger *slog.Logger) *LoggingDecorator {
	return &LoggingDecorator{next: next, logger: logger}
}

// Save logs the execution of the Save method and delegates the call to the next repository.
func (r *LoggingDecorator) Save(ctx context.Context, order domain.Order) error {
	start := time.Now()
	err := r.next.Save(ctx, order)
	r.logger.Info("repository.Save",
		"order_id", order.ID,
		"duration_ms", time.Since(start).Milliseconds(),
		"error", err,
	)
	return err
}

// FindByID logs the execution of the FindByID method and delegates the call to the next repository.
func (r *LoggingDecorator) FindByID(ctx context.Context, id string) (domain.Order, error) {
	start := time.Now()
	order, err := r.next.FindByID(ctx, id)
	r.logger.Info("repository.FindByID",
		"order_id", id,
		"found", err == nil,
		"duration_ms", time.Since(start).Milliseconds(),
		"error", err,
	)
	return order, err
}
