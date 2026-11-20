package server

import (
	"log"
	"log/slog"
	"net/http"

	http_middleware "github.com/tombenke/ports-and-adapters-example/internal/infrastructure/observability/http"

	"github.com/gorilla/mux"
	"github.com/tombenke/ports-and-adapters-example/internal/application"
	"github.com/tombenke/ports-and-adapters-example/internal/infrastructure/adapter/httphandler"
	"github.com/tombenke/ports-and-adapters-example/internal/infrastructure/adapter/idgen"
	"github.com/tombenke/ports-and-adapters-example/internal/infrastructure/adapter/memory"
	"github.com/tombenke/ports-and-adapters-example/internal/infrastructure/adapter/notify"
	loggingnotify "github.com/tombenke/ports-and-adapters-example/internal/infrastructure/observability/notify"
	"github.com/tombenke/ports-and-adapters-example/internal/infrastructure/observability/repository"
)

func StartServer() {
	// Create outbound adapters
	// e.g., repository, notifier
	repo := repository.NewLoggingDecorator(memory.NewOrderRepository(), slog.Default())
	notifier := loggingnotify.NewLoggingDecorator(notify.NewConsoleNotifier(), slog.Default())
	idgen := idgen.NewUUIDGenerator()

	// Create application service - inject dependencies
	svc := application.NewOrderService(repo, notifier, idgen)

	// Create inbound adapters
	// e.g., HTTP handlers, gRPC servers
	m := mux.NewRouter()
	handler := httphandler.New(svc, svc, svc, svc)
	handler.RegisterRoutes(m)

	// Wrap the entire mux with logging middleware
	logged := http_middleware.RequestLogging(slog.Default(), m)

	// Start the HTTP server with the handler
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", logged); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
