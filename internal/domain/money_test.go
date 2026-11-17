package domain_test

import (
	"testing"

	"github.com/tombenke/ports-and-adapters-example/internal/domain"
)

func TestMoney_Add_CurrencyMismatch(t *testing.T) {
	usd, _ := domain.NewMoney(1000, "USD")
	eur, _ := domain.NewMoney(500, "EUR")
	_, err := usd.Add(eur)
	if err == nil {
		t.Fatal("expected currency mismatch error")
	}
}
