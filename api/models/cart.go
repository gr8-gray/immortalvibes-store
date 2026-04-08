package models

// LineItem represents one product/price combination in a cart.
type LineItem struct {
	PriceID   string `json:"price_id"`   // Stripe Price ID
	ProductID string `json:"product_id"` // Stripe Product ID
	Name      string `json:"name"`
	ImageURL  string `json:"image_url"`
	Currency  string `json:"currency"`
	Amount    int64  `json:"amount"`     // unit price in smallest currency unit
	Quantity  int    `json:"quantity"`
}

// Cart is the full cart stored in Cloudflare KV.
type Cart struct {
	Token     string     `json:"token"`
	LineItems []LineItem `json:"line_items"`
}

// Total returns the sum of (Amount * Quantity) for all line items.
func (c *Cart) Total() int64 {
	var t int64
	for _, li := range c.LineItems {
		t += li.Amount * int64(li.Quantity)
	}
	return t
}
