package idgen

import (
	"github.com/google/uuid"
	"github.com/tombenke/ports-and-adapters-example/internal/application/ports"
)

// UUIDGenerator produces UUIDs for order IDs.
type UUIDGenerator struct{}

// Ensure UUIDGenerator implements the IDGenerator interface.
// This will cause a compile-time error if it doesn't.
// It is a compile-time assertion.
// See: https://golang.org/doc/faq#guarantee_satisfies_interface
var _ ports.IDGenerator = (*UUIDGenerator)(nil)

// NewUUIDGenerator creates a new instance of UUIDGenerator.
func NewUUIDGenerator() *UUIDGenerator { return &UUIDGenerator{} }

// NewID generates a new unique ID using UUID.
func (g *UUIDGenerator) NewID() string { return uuid.New().String() }
