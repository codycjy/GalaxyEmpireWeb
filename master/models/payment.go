package models

import (
	"time"

	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCanceled  PaymentStatus = "canceled"
	PaymentStatusRefunded  PaymentStatus = "refunded"
	PaymentStatusDisputed  PaymentStatus = "disputed"
)

type Payment struct {
	gorm.Model
	UserID         uint          `json:"user_id" gorm:"index;not null"`
	User           User          `gorm:"foreignKey:UserID"`
	Amount         int64         `json:"amount"`
	Currency       string        `json:"currency" gorm:"size:3"`
	Status         PaymentStatus `json:"status" gorm:"size:20"`
	PaymentIntent  string        `json:"payment_intent" gorm:"uniqueIndex;size:255"`
	SessionID      string        `json:"session_id" gorm:"uniqueIndex;size:255"`
	IdempotencyKey string        `json:"idempotency_key" gorm:"uniqueIndex;size:255"`
	PaidAt         time.Time     `json:"paid_at"`
	FailureReason  *string       `json:"failure_reason" gorm:"size:255"`
	RefundStatus   *string       `json:"refund_status" gorm:"size:20"`
	Description    string        `json:"description" gorm:"size:1000"`
	CustomerEmail  string        `json:"customer_email" gorm:"size:255"`
	Metadata       string        `json:"metadata" gorm:"type:text"`
	FulfilledAt    *time.Time    `json:"fulfilled_at"`
}
