package repository

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/tombenke/ports-and-adapters-example/internal/domain"
	"github.com/tombenke/ports-and-adapters-example/internal/infrastructure/adapter/memory"
)

func TestLoggingDecorator(t *testing.T) {
	inner := memory.NewOrderRepository()
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelInfo}))
	repo := NewLoggingDecorator(inner, logger)
	ctx := context.Background()

	price, err := domain.NewMoney(2000, "USD")
	if err != nil {
		t.Fatal(err)
	}
	order, err := domain.NewOrder("ord-1", "cust-1", []domain.LineItem{
		{ProductID: "p1", Quantity: 1, Price: price},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.Save(ctx, order); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := repo.FindByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.ID != order.ID || got.CustomerID != order.CustomerID || got.Status != order.Status {
		t.Errorf("FindByID() = %#v, want %#v", got, order)
	}

	_, err = repo.FindByID(ctx, "missing")
	if !errors.Is(err, domain.ErrOrderNotFound) {
		t.Errorf("FindByID(missing) error = %v, want ErrOrderNotFound", err)
	}

	for _, operation := range []string{"repository.Save", "repository.FindByID"} {
		if !strings.Contains(logOutput.String(), operation) {
			t.Errorf("log output does not contain %q: %s", operation, logOutput.String())
		}
	}
}
