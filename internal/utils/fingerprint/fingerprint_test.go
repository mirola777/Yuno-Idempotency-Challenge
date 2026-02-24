package fingerprint

import (
	"testing"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"github.com/stretchr/testify/assert"
)

func baseRequest() domain.PaymentRequest {
	return domain.PaymentRequest{
		Amount:     100.0,
		Currency:   domain.CurrencyIDR,
		CustomerID: "cust-1",
		RideID:     "ride-1",
		CardNumber: "4111111111111111",
	}
}

func TestCompute_ConsistentHash(t *testing.T) {
	req := baseRequest()
	hash1 := Compute(req)
	hash2 := Compute(req)

	assert.Equal(t, hash1, hash2)
}

func TestCompute_DifferentAmount(t *testing.T) {
	req1 := baseRequest()
	req2 := baseRequest()
	req2.Amount = 200.0

	assert.NotEqual(t, Compute(req1), Compute(req2))
}

func TestCompute_DifferentCurrency(t *testing.T) {
	req1 := baseRequest()
	req2 := baseRequest()
	req2.Currency = domain.CurrencyTHB

	assert.NotEqual(t, Compute(req1), Compute(req2))
}

func TestCompute_DifferentCustomerID(t *testing.T) {
	req1 := baseRequest()
	req2 := baseRequest()
	req2.CustomerID = "cust-2"

	assert.NotEqual(t, Compute(req1), Compute(req2))
}

func TestCompute_DifferentCardNumber(t *testing.T) {
	req1 := baseRequest()
	req2 := baseRequest()
	req2.CardNumber = "5500000000000004"

	assert.NotEqual(t, Compute(req1), Compute(req2))
}

func TestCompute_HashLength64(t *testing.T) {
	hash := Compute(baseRequest())

	assert.Len(t, hash, 64)
}
