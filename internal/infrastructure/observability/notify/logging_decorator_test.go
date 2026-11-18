package notify

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/tombenke/ports-and-adapters-example/internal/domain"
	consolenotify "github.com/tombenke/ports-and-adapters-example/internal/infrastructure/adapter/notify"
)

func TestLoggingDecorator_OrderPlaced(t *testing.T) {
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelInfo}))
	decorator := NewLoggingDecorator(consolenotify.NewConsoleNotifier(), logger)
	order := domain.Order{ID: "ord-1", CustomerID: "cust-1"}

	consoleOutput := captureStdout(t, func() {
		if err := decorator.OrderPlaced(context.Background(), order); err != nil {
			t.Fatalf("OrderPlaced() error = %v", err)
		}
	})
	if want := "[NOTIFICATION] Order ord-1 placed by customer cust-1\n"; consoleOutput != want {
		t.Errorf("console output = %q, want %q", consoleOutput, want)
	}
	for _, value := range []string{"notification.OrderPlaced", "order_id=ord-1", "customer_id=cust-1", "error=<nil>"} {
		if !strings.Contains(logOutput.String(), value) {
			t.Errorf("log output does not contain %q: %s", value, logOutput.String())
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previousStdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = previousStdout
		_ = writer.Close()
		_ = reader.Close()
	}()

	fn()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = previousStdout
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}
