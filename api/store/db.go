package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/immortalvibes/api/models"
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
	// Order manifest — the line items, persisted at checkout so the order is
	// self-contained (STOP 18). The cart in KV is deleted on payment.
	LineItems []models.LineItem
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
		ALTER TABLE orders ADD COLUMN IF NOT EXISTS line_items      JSONB;
	`)
	if err != nil {
		return err
	}
	// Subscribers — email list for discount codes.
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS subscribers (
			email      TEXT PRIMARY KEY,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return err
	}
	// Per-variant stock — additive, safe to run on every boot.
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS variant_stock (
			product_id  TEXT NOT NULL,
			variant     TEXT NOT NULL,
			stock_count INT  NOT NULL DEFAULT 0,
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (product_id, variant)
		);
	`)
	return err
}

// SaveSubscriber upserts an email into the subscribers table.
// Returns isNew=true if the row was freshly inserted.
func (d *DB) SaveSubscriber(ctx context.Context, emailAddr string) (bool, error) {
	result, err := d.db.ExecContext(ctx, `
		INSERT INTO subscribers (email) VALUES ($1)
		ON CONFLICT (email) DO NOTHING
	`, emailAddr)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// VariantStockRow holds a single per-variant stock record.
type VariantStockRow struct {
	Variant    string
	StockCount int
}

// GetVariantStocks returns all variant rows for a product, ordered by variant name.
// Returns an empty slice when no rows exist.
func (d *DB) GetVariantStocks(ctx context.Context, productID string) ([]VariantStockRow, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT variant, stock_count FROM variant_stock WHERE product_id = $1 ORDER BY variant`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VariantStockRow
	for rows.Next() {
		var r VariantStockRow
		if err := rows.Scan(&r.Variant, &r.StockCount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// SetVariantStock upserts a variant row for a product.
func (d *DB) SetVariantStock(ctx context.Context, productID, variant string, count int) error {
	_, err := d.db.ExecContext(ctx, `
		INSERT INTO variant_stock (product_id, variant, stock_count, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (product_id, variant) DO UPDATE
		SET stock_count = $3, updated_at = NOW()
	`, productID, variant, count)
	return err
}

// GetStock returns the current stock count for a Stripe Product ID.
// If variant rows exist, returns their SUM; otherwise falls back to legacy product_stock.
func (d *DB) GetStock(ctx context.Context, productID string) (int, error) {
	var sum sql.NullInt64
	err := d.db.QueryRowContext(ctx,
		`SELECT SUM(stock_count) FROM variant_stock WHERE product_id = $1`, productID,
	).Scan(&sum)
	if err != nil {
		return 0, err
	}
	if sum.Valid {
		return int(sum.Int64), nil
	}
	// No variant rows — fall back to legacy product_stock.
	var count int
	err = d.db.QueryRowContext(ctx,
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

// DecrementStock subtracts qty from stock atomically.
// If variant is non-empty, attempts the variant_stock row first.
// Falls back to legacy product_stock when no variant row matched (e.g. old carts).
// Returns ErrInsufficientStock if the result would go below zero.
func (d *DB) DecrementStock(ctx context.Context, productID, variant string, qty int) error {
	if variant != "" {
		res, err := d.db.ExecContext(ctx, `
			UPDATE variant_stock
			SET stock_count = stock_count - $3, updated_at = NOW()
			WHERE product_id = $1 AND variant = $2
			  AND stock_count >= $3
		`, productID, variant, qty)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows > 0 {
			log.Printf("store: DecrementStock product=%s variant=%s qty=%d — variant path", productID, variant, qty)
			return nil
		}
		// No variant row matched — fall through to legacy path.
		log.Printf("store: DecrementStock product=%s variant=%s qty=%d — no variant row, falling back to legacy", productID, variant, qty)
	}
	// Legacy product_stock path.
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
	log.Printf("store: DecrementStock product=%s qty=%d — legacy path", productID, qty)
	return nil
}

// SaveOrder inserts a new order row. PaymentIntentID must be unique.
// The order manifest (line items) is persisted as JSONB so the order is
// self-contained even after the cart is deleted on payment (STOP 18).
func (d *DB) SaveOrder(ctx context.Context, o OrderRow) error {
	items := o.LineItems
	if items == nil {
		items = []models.LineItem{}
	}
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshal line_items: %w", err)
	}
	_, err = d.db.ExecContext(ctx, `
		INSERT INTO orders (id, payment_intent_id, cart_token, email, currency, total_amount, status,
		                    shipping_name, line1, line2, city, state, postal_code, country, line_items, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW())
	`, o.ID, o.PaymentIntentID, o.CartToken, o.Email, o.Currency, o.TotalAmount, o.Status,
		o.ShippingName, o.Line1, o.Line2, o.City, o.State, o.PostalCode, o.Country, itemsJSON)
	return err
}

// scanLineItems unmarshals the JSONB line_items column (may be NULL/empty).
func scanLineItems(raw []byte, o *OrderRow) error {
	if len(raw) == 0 {
		o.LineItems = []models.LineItem{}
		return nil
	}
	return json.Unmarshal(raw, &o.LineItems)
}

// GetOrder retrieves an order by its UUID. Returns ErrOrderNotFound if missing.
func (d *DB) GetOrder(ctx context.Context, id string) (*OrderRow, error) {
	var o OrderRow
	// COALESCE nullable columns to '' — same NULL-scan bug as GetOrderByPaymentIntent
	// (incident 2026-06-21); GetOrder backs the /api/order/{id} confirmation page.
	var itemsRaw []byte
	err := d.db.QueryRowContext(ctx, `
		SELECT id, payment_intent_id, cart_token, email, currency, total_amount, status, created_at,
		       shipping_name, line1, COALESCE(line2,''), city, state, postal_code, country,
		       COALESCE(tracking_number,''), COALESCE(carrier,''), COALESCE(label_url,''), line_items
		FROM orders WHERE id = $1
	`, id).Scan(
		&o.ID, &o.PaymentIntentID, &o.CartToken, &o.Email, &o.Currency, &o.TotalAmount, &o.Status, &o.CreatedAt,
		&o.ShippingName, &o.Line1, &o.Line2, &o.City, &o.State, &o.PostalCode, &o.Country,
		&o.TrackingNumber, &o.Carrier, &o.LabelURL, &itemsRaw,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := scanLineItems(itemsRaw, &o); err != nil {
		return nil, fmt.Errorf("scan line_items: %w", err)
	}
	return &o, nil
}

// GetOrderByPaymentIntent retrieves an order by its Stripe PaymentIntent ID.
func (d *DB) GetOrderByPaymentIntent(ctx context.Context, paymentIntentID string) (*OrderRow, error) {
	var o OrderRow
	// COALESCE the nullable columns to '' — an unshipped order always has NULL
	// tracking_number/carrier/label_url, and Scan into a plain string fails on NULL.
	// This bug blocked EVERY order's fulfillment webhook (incident 2026-06-21).
	var itemsRaw []byte
	err := d.db.QueryRowContext(ctx, `
		SELECT id, payment_intent_id, cart_token, email, currency, total_amount, status, created_at,
		       shipping_name, line1, COALESCE(line2,''), city, state, postal_code, country,
		       COALESCE(tracking_number,''), COALESCE(carrier,''), COALESCE(label_url,''), line_items
		FROM orders WHERE payment_intent_id = $1
	`, paymentIntentID).Scan(
		&o.ID, &o.PaymentIntentID, &o.CartToken, &o.Email, &o.Currency, &o.TotalAmount, &o.Status, &o.CreatedAt,
		&o.ShippingName, &o.Line1, &o.Line2, &o.City, &o.State, &o.PostalCode, &o.Country,
		&o.TrackingNumber, &o.Carrier, &o.LabelURL, &itemsRaw,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := scanLineItems(itemsRaw, &o); err != nil {
		return nil, fmt.Errorf("scan line_items: %w", err)
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
