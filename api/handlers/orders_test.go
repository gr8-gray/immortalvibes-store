package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/immortalvibes/api/handlers"
	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/store"
)

// stubOrderStore implements handlers.OrderStore for tests.
type stubOrderStore struct {
	orders map[string]*store.OrderRow
}

func newStubOrderStore() *stubOrderStore {
	return &stubOrderStore{orders: map[string]*store.OrderRow{}}
}

func (s *stubOrderStore) GetOrder(ctx context.Context, id string) (*store.OrderRow, error) {
	o, ok := s.orders[id]
	if !ok {
		return nil, store.ErrOrderNotFound
	}
	return o, nil
}

func (s *stubOrderStore) UpdateOrderShipping(ctx context.Context, id, trackingNumber, carrier, labelURL string) error {
	if o, ok := s.orders[id]; ok {
		o.TrackingNumber = trackingNumber
		o.Carrier = carrier
		o.LabelURL = labelURL
	}
	return nil
}

func TestGetOrder(t *testing.T) {
	ss := newStubOrderStore()
	ss.orders["ord-001"] = &store.OrderRow{
		ID:              "ord-001",
		PaymentIntentID: "pi_test",
		CartToken:       "tok-1",
		Email:           "user@example.com",
		Currency:        "gbp",
		TotalAmount:     5999,
		Status:          "complete",
		CreatedAt:       time.Now(),
	}

	h := handlers.NewOrdersHandler(ss)
	r := chi.NewRouter()
	r.Get("/api/order/{id}", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/api/order/ord-001", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var order models.Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if order.ID != "ord-001" {
		t.Errorf("ID = %q, want ord-001", order.ID)
	}
	if order.Email != "user@example.com" {
		t.Errorf("Email = %q, want user@example.com", order.Email)
	}
	if order.TotalAmount != 5999 {
		t.Errorf("TotalAmount = %d, want 5999", order.TotalAmount)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	ss := newStubOrderStore()
	h := handlers.NewOrdersHandler(ss)
	r := chi.NewRouter()
	r.Get("/api/order/{id}", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/api/order/no-such", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}
