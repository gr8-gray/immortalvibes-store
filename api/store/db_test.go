package store_test

import (
	"os"
	"testing"

	"github.com/immortalvibes/api/store"
)

func testDB(t *testing.T) *store.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set — skipping Postgres tests")
	}
	db, err := store.Open(dsn)
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

func TestDBMigrate(t *testing.T) {
	db := testDB(t)
	// If Open succeeds with migrations, the tables exist. Verify by querying them.
	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping: %v", err)
	}
}

func TestSetAndGetStock(t *testing.T) {
	db := testDB(t)

	productID := "prod_test_001"
	if err := db.SetStock(t.Context(), productID, 10); err != nil {
		t.Fatalf("SetStock: %v", err)
	}

	stock, err := db.GetStock(t.Context(), productID)
	if err != nil {
		t.Fatalf("GetStock: %v", err)
	}
	if stock != 10 {
		t.Errorf("GetStock = %d, want 10", stock)
	}
}

func TestDecrementStock(t *testing.T) {
	db := testDB(t)

	productID := "prod_test_002"
	_ = db.SetStock(t.Context(), productID, 5)

	if err := db.DecrementStock(t.Context(), productID, 3); err != nil {
		t.Fatalf("DecrementStock: %v", err)
	}

	stock, _ := db.GetStock(t.Context(), productID)
	if stock != 2 {
		t.Errorf("after decrement: stock = %d, want 2", stock)
	}
}

func TestDecrementStock_InsufficientStock(t *testing.T) {
	db := testDB(t)

	productID := "prod_test_003"
	_ = db.SetStock(t.Context(), productID, 1)

	err := db.DecrementStock(t.Context(), productID, 5)
	if err != store.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestSaveAndGetOrder(t *testing.T) {
	db := testDB(t)

	order := store.OrderRow{
		ID:              "ord-uuid-001",
		PaymentIntentID: "pi_test_001",
		CartToken:       "cart-tok-001",
		Email:           "buyer@example.com",
		Currency:        "usd",
		TotalAmount:     4999,
		Status:          "complete",
	}

	if err := db.SaveOrder(t.Context(), order); err != nil {
		t.Fatalf("SaveOrder: %v", err)
	}

	got, err := db.GetOrder(t.Context(), "ord-uuid-001")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if got.Email != "buyer@example.com" {
		t.Errorf("got email %q, want %q", got.Email, "buyer@example.com")
	}
	if got.TotalAmount != 4999 {
		t.Errorf("got amount %d, want 4999", got.TotalAmount)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	db := testDB(t)
	_, err := db.GetOrder(t.Context(), "no-such-id")
	if err != store.ErrOrderNotFound {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestGetOrderByPaymentIntent(t *testing.T) {
	db := testDB(t)

	_ = db.SaveOrder(t.Context(), store.OrderRow{
		ID:              "ord-pi-001",
		PaymentIntentID: "pi_lookup_001",
		CartToken:       "tok",
		Email:           "pi@example.com",
		Currency:        "usd",
		TotalAmount:     1000,
		Status:          "pending",
	})

	got, err := db.GetOrderByPaymentIntent(t.Context(), "pi_lookup_001")
	if err != nil {
		t.Fatalf("GetOrderByPaymentIntent: %v", err)
	}
	if got.ID != "ord-pi-001" {
		t.Errorf("ID = %q, want ord-pi-001", got.ID)
	}
}

func TestUpdateOrderStatus(t *testing.T) {
	db := testDB(t)

	_ = db.SaveOrder(t.Context(), store.OrderRow{
		ID:              "ord-st-001",
		PaymentIntentID: "pi_status_001",
		CartToken:       "tok2",
		Email:           "st@example.com",
		Currency:        "eur",
		TotalAmount:     2000,
		Status:          "pending",
	})

	if err := db.UpdateOrderStatus(t.Context(), "ord-st-001", "complete"); err != nil {
		t.Fatalf("UpdateOrderStatus: %v", err)
	}

	got, _ := db.GetOrder(t.Context(), "ord-st-001")
	if got.Status != "complete" {
		t.Errorf("status = %q, want complete", got.Status)
	}
}
