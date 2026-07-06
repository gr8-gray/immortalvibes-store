package models

// Price holds a single currency/amount pair for a product.
type Price struct {
	PriceID  string `json:"price_id"`  // Stripe Price ID
	Currency string `json:"currency"`  // lowercase ISO: "usd", "gbp", "eur", "aud"
	Amount   int64  `json:"amount"`    // smallest currency unit (cents/pence)
}

// VariantStock holds per-variant (size/colorway) stock for a product.
type VariantStock struct {
	Variant    string `json:"variant"`     // e.g. "L", "Earth Blue"
	StockCount int    `json:"stock_count"`
}

// Product is the enriched product returned to the frontend.
type Product struct {
	ID          string         `json:"id"`           // Stripe Product ID
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ImageURL    string         `json:"image_url"`    // R2 public URL
	Prices      []Price        `json:"prices"`
	StockCount  int            `json:"stock_count"`
	Variants    []VariantStock `json:"variants"`     // empty array = no per-variant data
}
