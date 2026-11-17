package dto

import "github.com/tombenke/ports-and-adapters-example/internal/domain"

// PlaceOrderRequest represents the input data required to place a new order.
type PlaceOrderRequest struct {
	CustomerID string
	Items      []domain.LineItem
}
