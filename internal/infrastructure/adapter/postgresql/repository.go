package postgresql

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tombenke/ports-and-adapters-example/internal/application/ports"
	"github.com/tombenke/ports-and-adapters-example/internal/domain"
)

//go:embed schema.sql
var schema string

var (
	_ ports.OrderRepository = (*OrderRepository)(nil)
	_ ports.OrderLister     = (*OrderRepository)(nil)
)

// OrderRepository stores orders in PostgreSQL using a caller-owned connection pool.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a repository without opening or closing the database.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// EnsureSchema creates the repository table if it does not exist.
func (r *OrderRepository) EnsureSchema(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("creating order schema: %w", err)
	}
	return nil
}

type storedItem struct {
	ProductID  string `json:"product_id"`
	Quantity   int    `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
	Currency   string `json:"currency"`
}

// Save atomically inserts or replaces the entire order with the same ID.
func (r *OrderRepository) Save(ctx context.Context, order domain.Order) error {
	var items []storedItem
	if order.Items != nil {
		items = make([]storedItem, len(order.Items))
	}
	for index, item := range order.Items {
		items[index] = storedItem{
			ProductID: item.ProductID, Quantity: item.Quantity,
			PriceCents: item.Price.Amount(), Currency: item.Price.Currency(),
		}
	}
	data, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("encoding order items: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO orders (id, customer_id, status, created_at, items)
		VALUES ($1, $2, $3, $4, $5::jsonb)
		ON CONFLICT (id) DO UPDATE SET
			customer_id = EXCLUDED.customer_id,
			status = EXCLUDED.status,
			created_at = EXCLUDED.created_at,
			items = EXCLUDED.items`,
		order.ID, order.CustomerID, string(order.Status), order.CreatedAt, string(data))
	if err != nil {
		return fmt.Errorf("saving order %q: %w", order.ID, err)
	}
	return nil
}

// FindByID returns domain.ErrOrderNotFound when the order does not exist.
func (r *OrderRepository) FindByID(ctx context.Context, id string) (domain.Order, error) {
	order, err := scanOrder(r.db.QueryRowContext(ctx,
		`SELECT id, customer_id, status, created_at, items FROM orders WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("finding order %q: %w", id, err)
	}
	return order, nil
}

// ListByCustomer returns all matching orders, ordered by creation time and ID.
func (r *OrderRepository) ListByCustomer(ctx context.Context, customerID string) ([]domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, customer_id, status, created_at, items FROM orders
		WHERE customer_id = $1 ORDER BY created_at, id`, customerID)
	if err != nil {
		return nil, fmt.Errorf("listing customer orders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var orders []domain.Order
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("reading customer order: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating customer orders: %w", err)
	}
	return orders, nil
}

func scanOrder(row interface{ Scan(...any) error }) (domain.Order, error) {
	var order domain.Order
	var data []byte
	if err := row.Scan(&order.ID, &order.CustomerID, &order.Status, &order.CreatedAt, &data); err != nil {
		return domain.Order{}, err
	}
	var items []storedItem
	if err := json.Unmarshal(data, &items); err != nil {
		return domain.Order{}, fmt.Errorf("decoding order items: %w", err)
	}
	if items != nil {
		order.Items = make([]domain.LineItem, len(items))
	}
	for index, item := range items {
		price, err := domain.NewMoney(item.PriceCents, item.Currency)
		if err != nil {
			return domain.Order{}, fmt.Errorf("decoding order item %d price: %w", index, err)
		}
		order.Items[index] = domain.LineItem{
			ProductID: item.ProductID, Quantity: item.Quantity, Price: price,
		}
	}
	return order, nil
}
