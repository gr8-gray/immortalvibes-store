package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var ErrInsufficientStock = errors.New("insufficient stock")
var ErrOrderNotFound = errors.New("order not found")

// DB wraps *sql.DB and exposes domain-level methods.
type DB struct {
	db *sql.DB
}

// OrderRow is the flat struct used for DB reads and writes.
type OrderRow struct {
	ID              string
	PaymentIntentID string
	CartToken       string
	Email           string
	Currency        string
	TotalAmount     int64
	Status          string
	CreatedAt       time.Time
	// Shipping address — collected at checkout
	ShippingName string
	Line1        string
	Line2        string
	City         string
	State        string
	PostalCode   string
	Country      string
	// Set by webhook after label purchase
	TrackingNumber string
	Carrier        string
	LabelURL       string
}

// Open connects to Postgres and runs migrations. Returns a ready-to-use DB.
func Open(dsn string) (*DB, error) {
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	d := &DB{db: sqlDB}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

// Ping delegates to the underlying sql.DB.
func (d *DB) Ping() error {
	return d.db.Ping()
}

// Close closes the underlying connection pool.
func (d *DB) Close() {
	d.db.Close()
}

func (d *DB) migrate() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS product_stock (
			product_id  TEXT PRIMARY KEY,
			stock_count INT  NOT NULL DEFAULT 0,
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS orders (
			id                TEXT PRIMARY KEY,
			payment_intent_id TEXT NOT NULL UNIQUE,
			cart_token        TEXT NOT NULL,
			email             TEXT NOT NULL,
			currency          TEXT NOT NULL,
			total_amount      BIGINT NOT NULL,
			status            TEXT NOT NULL DEFAULT 'pending',
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return err
	}
	// Shipping columns — idempotent, safe to run on every boot.
	_, err = d.db.Exec(`
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_name   TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS line1           TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS line2           TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS city            TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS state           TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS postal_code     TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS country         TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS tracking_number TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS carrier         TEXT;
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS label_url       TEXT;
	`)
	return err
}

// GetStock returns the current stock count for a Stripe Product ID.
// Returns 0 if the product has no stock row.
func (d *DB) GetStock(ctx context.Context, productID string) (int, error) {
	var count int
	err := d.db.QueryRowContext(ctx,
		`SELECT stock_count FROM product_stock WHERE product_id = $1`, productID,
	).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return count, err
}

// SetStock upserts the stock count for a product.
func (d *DB) SetStock(ctx context.Context, productID string, count int) error {
	_, err := d.db.ExecContext(ctx, `
		INSERT INTO product_stock (product_id, stock_count, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (product_id) DO UPDATE
		SET stock_count = $2, updated_at = NOW()
	`, productID, count)
	return err
}

// DecrementStock subtracts qty from product stock atomically.
// Returns ErrInsufficientStock if the result would go below zero.
func (d *DB) DecrementStock(ctx context.Context, productID string, qty int) error {
	res, err := d.db.ExecContext(ctx, `
		UPDATE product_stock
		SET stock_count = stock_count - $2, updated_at = NOW()
		WHERE product_id = $1
		  AND stock_count >= $2
	`, productID, qty)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInsufficientStock
	}
	return nil
}

// SaveOrder inserts a new order row. PaymentIntentID must be unique.
func (d *DB) SaveOrder(ctx context.Context, o OrderRow) error {
	_, err := d.db.ExecContext(ctx, `
		INSERT INTO orders (id, payment_intent_id, cart_token, email, currency, total_amount, status,
		                    shipping_name, line1, line2, city, state, postal_code, country, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
	`, o.ID, o.PaymentIntentID, o.CartToken, o.Email, o.Currency, o.TotalAmount, o.Status,
		o.ShippingName, o.Line1, o.Line2, o.City, o.State, o.PostalCode, o.Country)
	return err
}

// GetOrder retrieves an order by its UUID. Returns ErrOrderNotFound if missing.
func (d *DB) GetOrder(ctx context.Context, id string) (*OrderRow, error) {
	var o OrderRow
	err := d.db.QueryRowContext(ctx, `
		SELECT id, payment_intent_id, cart_token, email, currency, total_amount, status, created_at,
		       shipping_name, line1, line2, city, state, postal_code, country,
		       tracking_number, carrier, label_url
		FROM orders WHERE id = $1
	`, id).Scan(
		&o.ID, &o.PaymentIntentID, &o.CartToken, &o.Email, &o.Currency, &o.TotalAmount, &o.Status, &o.CreatedAt,
		&o.ShippingName, &o.Line1, &o.Line2, &o.City, &o.State, &o.PostalCode, &o.Country,
		&o.TrackingNumber, &o.Carrier, &o.LabelURL,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// GetOrderByPaymentIntent retrieves an order by its Stripe PaymentIntent ID.
func (d *DB) GetOrderByPaymentIntent(ctx context.Context, paymentIntentID string) (*OrderRow, error) {
	var o OrderRow
	err := d.db.QueryRowContext(ctx, `
		SELECT id, payment_intent_id, cart_token, email, currency, total_amount, status, created_at,
		       shipping_name, line1, line2, city, state, postal_code, country,
		       tracking_number, carrier, label_url
		FROM orders WHERE payment_intent_id = $1
	`, paymentIntentID).Scan(
		&o.ID, &o.PaymentIntentID, &o.CartToken, &o.Email, &o.Currency, &o.TotalAmount, &o.Status, &o.CreatedAt,
		&o.ShippingName, &o.Line1, &o.Line2, &o.City, &o.State, &o.PostalCode, &o.Country,
		&o.TrackingNumber, &o.Carrier, &o.LabelURL,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// UpdateOrderShipping sets the tracking and label fields after a label is purchased.
func (d *DB) UpdateOrderShipping(ctx context.Context, id, trackingNumber, carrier, labelURL string) error {
	_, err := d.db.ExecContext(ctx, `
		UPDATE orders SET tracking_number=$2, carrier=$3, label_url=$4 WHERE id=$1
	`, id, trackingNumber, carrier, labelURL)
	return err
}

// UpdateOrderStatus sets the status field for an order by ID.
func (d *DB) UpdateOrderStatus(ctx context.Context, id, status string) error {
	_, err := d.db.ExecContext(ctx, `
		UPDATE orders SET status = $2 WHERE id = $1
	`, id, status)
	return err
}
