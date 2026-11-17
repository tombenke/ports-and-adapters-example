package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tombenke/ports-and-adapters-example/internal/application/dto"
	"github.com/tombenke/ports-and-adapters-example/internal/application/ports"
	"github.com/tombenke/ports-and-adapters-example/internal/domain"
)

// Handler is the HTTP handler for order-related endpoints.
// It implements the necessary routes for creating, retrieving, confirming, and canceling orders.
type Handler struct {
	placer    ports.OrderPlacer
	getter    ports.OrderGetter
	confirmer ports.OrderConfirmer
	canceler  ports.OrderCanceler
}

func New(p ports.OrderPlacer, g ports.OrderGetter, c ports.OrderConfirmer, ca ports.OrderCanceler) *Handler {
	return &Handler{
		placer:    p,
		getter:    g,
		confirmer: c,
		canceler:  ca,
	}
}

func (h *Handler) RegisterRoutes(m *mux.Router) {
	m.HandleFunc("/orders", h.CreateOrder()).Methods("POST")
	m.HandleFunc("/orders/{id}", h.GetOrder()).Methods("GET")
	m.HandleFunc("/orders/{id}/confirm", h.ConfirmOrder()).Methods("POST")
	m.HandleFunc("/orders/{id}/cancel", h.CancelOrder()).Methods("POST")
}

// --- Request/Response DTOs ---

type createOrderRequest struct {
	CustomerID string            `json:"customer_id"`
	Items      []createOrderItem `json:"items"`
}

type createOrderItem struct {
	ProductID  string `json:"product_id"`
	Quantity   int    `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
	Currency   string `json:"currency"`
}

type orderResponse struct {
	ID         string              `json:"id"`
	CustomerID string              `json:"customer_id"`
	Items      []orderItemResponse `json:"items"`
	Status     string              `json:"status"`
	CreatedAt  string              `json:"created_at"`
}

type orderItemResponse struct {
	ProductID  string `json:"product_id"`
	Quantity   int    `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
	Currency   string `json:"currency"`
}

func toOrderResponse(order domain.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, orderItemResponse{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			PriceCents: item.Price.Amount(),
			Currency:   item.Price.Currency(),
		})
	}
	return orderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Items:      items,
		Status:     string(order.Status),
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- Handlers ---

func (h *Handler) CreateOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		items := make([]domain.LineItem, 0, len(req.Items))
		for _, item := range req.Items {
			currency := item.Currency
			if currency == "" {
				currency = "USD"
			}
			price, err := domain.NewMoney(item.PriceCents, currency)
			if err != nil {
				writeError(w, err.Error(), http.StatusBadRequest)
				return
			}
			items = append(items, domain.LineItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     price,
			})
		}

		order, err := h.placer.PlaceOrder(r.Context(), dto.PlaceOrderRequest{
			CustomerID: req.CustomerID,
			Items:      items,
		})
		if err != nil {
			writeError(w, err.Error(), domainErrorToStatus(err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(toOrderResponse(order)); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) GetOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		order, err := h.getter.GetOrder(r.Context(), id)
		if err != nil {
			writeError(w, err.Error(), domainErrorToStatus(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(toOrderResponse(order)); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) ConfirmOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		if err := h.confirmer.ConfirmOrder(r.Context(), id); err != nil {
			writeError(w, err.Error(), domainErrorToStatus(err))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) CancelOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		if err := h.canceler.CancelOrder(r.Context(), id); err != nil {
			writeError(w, err.Error(), domainErrorToStatus(err))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- Error mapping ---

func domainErrorToStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrOrderEmpty),
		errors.Is(err, domain.ErrNegativeAmount),
		errors.Is(err, domain.ErrInvalidDiscount):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrOrderNotPending),
		errors.Is(err, domain.ErrOrderCanceled):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, map[string]string{"error": msg})
}
