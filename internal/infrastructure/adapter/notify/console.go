package notify

import (
	"context"
	"fmt"

	"github.com/tombenke/ports-and-adapters-example/internal/application/ports"
	"github.com/tombenke/ports-and-adapters-example/internal/domain"
)

// Ensure ConsoleNotifier implements the ports.NotificationSender interface.
var _ ports.NotificationSender = (*ConsoleNotifier)(nil)

// ConsoleNotifier is an implementation of the ports.NotificationSender interface that prints notifications to the console.
type ConsoleNotifier struct{}

// NewConsoleNotifier creates a new instance of ConsoleNotifier.
func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

// OrderPlaced sends a notification that an order has been placed.
func (n *ConsoleNotifier) OrderPlaced(_ context.Context, order domain.Order) error {
	fmt.Printf("[NOTIFICATION] Order %s placed by customer %s\n",
		order.ID, order.CustomerID)
	return nil
}
