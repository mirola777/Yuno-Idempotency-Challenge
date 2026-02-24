package fingerprint

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
)

func Compute(req domain.PaymentRequest) string {
	data, _ := json.Marshal(req)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
