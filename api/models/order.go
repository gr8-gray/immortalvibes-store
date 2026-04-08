package models

import "time"

// Order is stored in Postgres after a successful PaymentIntent.
type Order struct {
	ID              string    `json:"id"`               // UUID
	PaymentIntentID string    `json:"payment_intent_id"`
	CartToken       string    `json:"cart_token"`
	Email           string    `json:"email"`
	Currency        string    `json:"currency"`
	TotalAmount     int64     `json:"total_amount"`
	Status          string    `json:"status"`            // "pending" | "complete"
	CreatedAt       time.Time `json:"created_at"`
}
