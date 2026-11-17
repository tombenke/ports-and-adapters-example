package application

import (
	"context"
	"fmt"

	"github.com/tombenke/ports-and-adapters-example/internal/application/dto"
	"github.com/tombenke/ports-and-adapters-example/internal/application/ports"
	"github.com/tombenke/ports-and-adapters-example/internal/domain"
)

// OrderService orchestrates the domain operations.
// It depends on ports (interfaces), NEVER on concrete adapters.
type OrderService struct {
	repo     ports.OrderRepository
	notifier ports.NotificationSender
	idgen    ports.IDGenerator
}

// NewOrderService creates a new instance of OrderService with the given dependencies.
func NewOrderService(repo ports.OrderRepository, notifier ports.NotificationSender, idgen ports.IDGenerator) *OrderService {
	return &OrderService{
		repo:     repo,
		notifier: notifier,
		idgen:    idgen,
	}
}

// PlaceOrder handles the business logic for placing a new order.
func (s *OrderService) PlaceOrder(ctx context.Context, req dto.PlaceOrderRequest) (domain.Order, error) {

	// Create a new order with a generated ID and the provided customer ID and items.
	order, err := domain.NewOrder(s.idgen.NewID(), req.CustomerID, req.Items)
	if err != nil {
		return domain.Order{}, fmt.Errorf("creating order: %w", err)
	}

	// Persist the new order using the repository.
	if err := s.repo.Save(ctx, order); err != nil {
		return domain.Order{}, fmt.Errorf("saving order: %w", err)
	}

	// Notify interested parties about the new order.
	// Notification failure shouldn't fail the order placement.
	_ = s.notifier.OrderPlaced(ctx, order)

	return order, nil
}

// GetOrder retrieves an order by its ID. Returns an error if the order is not found.
func (s *OrderService) GetOrder(ctx context.Context, id string) (domain.Order, error) {
	// TODO: Validate the input ID.
	// if id == "" {
	// 	return domain.Order{}, fmt.Errorf("invalid order ID")
	// }

	// Retrieve the order from the repository by its ID.
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

// ConfirmOrder confirms an existing order by its ID.
func (s *OrderService) ConfirmOrder(ctx context.Context, id string) error {
	// TODO: Validate the input ID.
	// if id == "" {
	// 	return fmt.Errorf("invalid order ID")
	// }

	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := order.Confirm(); err != nil {
		return err
	}
	return s.repo.Save(ctx, order)
}

// CancelOrder cancels an existing order by its ID.
func (s *OrderService) CancelOrder(ctx context.Context, id string) error {
	// TODO: Validate the input ID.
	// if id == "" {
	// 	return fmt.Errorf("invalid order ID")
	// }

	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := order.Cancel(); err != nil {
		return err
	}
	return s.repo.Save(ctx, order)
}
