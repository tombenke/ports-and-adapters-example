package postgresql_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/tombenke/ports-and-adapters-example/internal/domain"
	"github.com/tombenke/ports-and-adapters-example/internal/infrastructure/adapter/postgresql"
)

func TestOrderRepository(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("set POSTGRES_TEST_DSN to run PostgreSQL repository integration tests")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	schemaName := pq.QuoteIdentifier(fmt.Sprintf("order_repository_test_%d", time.Now().UnixNano()))
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := db.ExecContext(cleanupCtx, "DROP SCHEMA "+schemaName+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.ExecContext(ctx, "SET search_path TO "+schemaName); err != nil {
		t.Fatal(err)
	}
	repo := postgresql.NewOrderRepository(db)
	for range 2 {
		if err := repo.EnsureSchema(ctx); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("lifecycle and full replacement", func(t *testing.T) {
		order := testOrder(t, "lifecycle", "customer-1")
		for _, status := range []domain.OrderStatus{
			domain.OrderStatusPending, domain.OrderStatusConfirmed, domain.OrderStatusCanceled,
		} {
			order.Status = status
			if err := repo.Save(ctx, order); err != nil {
				t.Fatal(err)
			}
			got, err := repo.FindByID(ctx, order.ID)
			if err != nil {
				t.Fatal(err)
			}
			assertOrder(t, order, got)
		}
		replacement := testOrder(t, order.ID, "replacement-customer")
		replacement.Items = replacement.Items[1:]
		replacement.Items[0].Quantity = 7
		replacement.CreatedAt = replacement.CreatedAt.Add(time.Hour)
		if err := repo.Save(ctx, replacement); err != nil {
			t.Fatal(err)
		}
		got, err := repo.FindByID(ctx, order.ID)
		if err != nil {
			t.Fatal(err)
		}
		assertOrder(t, replacement, got)
	})

	t.Run("missing order", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "missing")
		if !errors.Is(err, domain.ErrOrderNotFound) {
			t.Fatalf("got %v, want ErrOrderNotFound", err)
		}
	})

	t.Run("nil and empty items", func(t *testing.T) {
		for name, items := range map[string][]domain.LineItem{"nil": nil, "empty": {}} {
			order := testOrder(t, name, "empty-items-customer")
			order.Items = items
			if err := repo.Save(ctx, order); err != nil {
				t.Fatal(err)
			}
			got, err := repo.FindByID(ctx, order.ID)
			if err != nil {
				t.Fatal(err)
			}
			assertOrder(t, order, got)
		}
	})

	t.Run("customer filtering", func(t *testing.T) {
		first := testOrder(t, "first", "filtered-customer")
		second := testOrder(t, "second", first.CustomerID)
		second.CreatedAt = first.CreatedAt.Add(time.Minute)
		for _, order := range []domain.Order{second, testOrder(t, "other", "other-customer"), first} {
			if err := repo.Save(ctx, order); err != nil {
				t.Fatal(err)
			}
		}
		orders, err := repo.ListByCustomer(ctx, first.CustomerID)
		if err != nil {
			t.Fatal(err)
		}
		if len(orders) != 2 {
			t.Fatalf("got %d orders, want 2", len(orders))
		}
		assertOrder(t, first, orders[0])
		assertOrder(t, second, orders[1])
		empty, err := repo.ListByCustomer(ctx, "missing-customer")
		if err != nil || len(empty) != 0 {
			t.Fatalf("got %v, %v; want no orders and no error", empty, err)
		}
	})

	t.Run("canceled context", func(t *testing.T) {
		canceledCtx, cancelContext := context.WithCancel(ctx)
		cancelContext()
		checks := map[string]func() error{
			"schema": func() error { return repo.EnsureSchema(canceledCtx) },
			"save":   func() error { return repo.Save(canceledCtx, testOrder(t, "canceled", "customer")) },
			"find": func() error {
				_, err := repo.FindByID(canceledCtx, "lifecycle")
				return err
			},
			"list": func() error {
				_, err := repo.ListByCustomer(canceledCtx, "customer")
				return err
			},
		}
		for name, check := range checks {
			if err := check(); !errors.Is(err, context.Canceled) {
				t.Errorf("%s: got %v, want context.Canceled", name, err)
			}
		}
	})
}

func testOrder(t *testing.T, id, customerID string) domain.Order {
	t.Helper()
	usd, err := domain.NewMoney(12345, "USD")
	if err != nil {
		t.Fatal(err)
	}
	eur, err := domain.NewMoney(987, "EUR")
	if err != nil {
		t.Fatal(err)
	}
	return domain.Order{
		ID: id, CustomerID: customerID, Status: domain.OrderStatusPending,
		CreatedAt: time.Date(2026, time.November, 18, 12, 0, 0, 123456000, time.UTC),
		Items: []domain.LineItem{
			{ProductID: "product-1", Quantity: 2, Price: usd},
			{ProductID: "product-2", Quantity: 3, Price: eur},
		},
	}
}

func assertOrder(t *testing.T, want, got domain.Order) {
	t.Helper()
	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("created_at = %v, want %v", got.CreatedAt, want.CreatedAt)
	}
	got.CreatedAt = want.CreatedAt
	if !reflect.DeepEqual(got, want) {
		t.Errorf("order = %#v, want %#v", got, want)
	}
}
