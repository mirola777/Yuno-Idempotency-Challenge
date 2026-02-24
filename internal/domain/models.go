package domain

import "time"

type PaymentStatus string

const (
	PaymentStatusSucceeded PaymentStatus = "SUCCEEDED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusPending   PaymentStatus = "PENDING"
)

type Currency string

const (
	CurrencyIDR Currency = "IDR"
	CurrencyTHB Currency = "THB"
	CurrencyVND Currency = "VND"
	CurrencyPHP Currency = "PHP"
)

var ValidCurrencies = map[Currency]bool{
	CurrencyIDR: true,
	CurrencyTHB: true,
	CurrencyVND: true,
	CurrencyPHP: true,
}

type IdempotencyStatus string

const (
	IdempotencyStatusProcessing IdempotencyStatus = "PROCESSING"
	IdempotencyStatusCompleted  IdempotencyStatus = "COMPLETED"
)

type PaymentRequest struct {
	Amount      float64  `json:"amount"`
	Currency    Currency `json:"currency"`
	CustomerID  string   `json:"customer_id"`
	RideID      string   `json:"ride_id"`
	CardNumber  string   `json:"card_number"`
	Description string   `json:"description,omitempty"`
}

type Payment struct {
	ID          string        `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Amount      float64       `json:"amount" gorm:"not null"`
	Currency    Currency      `json:"currency" gorm:"type:varchar(3);not null"`
	CustomerID  string        `json:"customer_id" gorm:"type:varchar(100);not null"`
	RideID      string        `json:"ride_id" gorm:"type:varchar(100);not null"`
	Status      PaymentStatus `json:"status" gorm:"type:varchar(20);not null"`
	CardLast4   string        `json:"card_last_4" gorm:"type:varchar(4)"`
	Description string        `json:"description,omitempty" gorm:"type:text"`
	FailReason  string        `json:"fail_reason,omitempty" gorm:"type:text"`
	CreatedAt   time.Time     `json:"created_at" gorm:"autoCreateTime"`
}

type IdempotencyRecord struct {
	Key                string            `json:"key" gorm:"primaryKey;type:varchar(64)"`
	RequestFingerprint string            `json:"request_fingerprint" gorm:"type:varchar(64);not null"`
	PaymentID          string            `json:"payment_id,omitempty" gorm:"type:varchar(36)"`
	ResponseBody       []byte            `json:"-" gorm:"type:jsonb"`
	Status             IdempotencyStatus `json:"status" gorm:"type:varchar(20);not null"`
	CreatedAt          time.Time         `json:"created_at" gorm:"autoCreateTime"`
	ExpiresAt          time.Time         `json:"expires_at" gorm:"index;not null"`
}

func (Payment) TableName() string {
	return "payments"
}

func (IdempotencyRecord) TableName() string {
	return "idempotency_records"
}
