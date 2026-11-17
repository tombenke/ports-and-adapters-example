package memory

import (
	"context"
	"sync"

	"github.com/tombenke/ports-and-adapters-example/internal/application/ports"
	"github.com/tombenke/ports-and-adapters-example/internal/domain"
)

// Ensure OrderRepository implements the domain.OrderRepository interface.
var _ ports.OrderRepository = (*OrderRepository)(nil)

// OrderRepository is an in-memory implementation of the domain.OrderRepository interface.
type OrderRepository struct {
	mu     sync.RWMutex
	orders map[string]domain.Order
}

// NewOrderRepository creates a new instance of OrderRepository.
func NewOrderRepository() *OrderRepository {
	return &OrderRepository{orders: make(map[string]domain.Order)}
}

// Save stores the given order in the repository.
// If an order with the same ID already exists, it will be overwritten.
func (r *OrderRepository) Save(_ context.Context, order domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
	return nil
}

// FindByID retrieves an order by its ID.
// Returns domain.ErrOrderNotFound if the order does not exist.
func (r *OrderRepository) FindByID(_ context.Context, id string) (domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, ok := r.orders[id]
	if !ok {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return order, nil
}

// ListByCustomer retrieves all orders for a given customer ID.
func (r *OrderRepository) ListByCustomer(_ context.Context, customerID string) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Order
	for _, o := range r.orders {
		if o.CustomerID == customerID {
			result = append(result, o)
		}
	}
	return result, nil
}
