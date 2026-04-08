package store_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/store"
)

func newKVServer(t *testing.T) (*httptest.Server, *store.KVClient) {
	t.Helper()
	storage := map[string]string{}

	mux := http.NewServeMux()

	// PUT value
	mux.HandleFunc("/client/v4/accounts/acct/storage/kv/namespaces/ns/values/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/client/v4/accounts/acct/storage/kv/namespaces/ns/values/")
		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			storage[key] = string(body)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
			return
		}
		if r.Method == http.MethodGet {
			v, ok := storage[key]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Write([]byte(v))
			return
		}
		if r.Method == http.MethodDelete {
			delete(storage, key)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
			return
		}
	})

	srv := httptest.NewServer(mux)
	client := store.NewKVClient(srv.URL, "acct", "ns", "token")
	return srv, client
}

func TestKVSetAndGetCart(t *testing.T) {
	srv, client := newKVServer(t)
	defer srv.Close()

	cart := &models.Cart{
		Token: "abc123",
		LineItems: []models.LineItem{
			{PriceID: "price_1", ProductID: "prod_1", Name: "T-Shirt", Currency: "usd", Amount: 2500, Quantity: 2},
		},
	}

	if err := client.SetCart(t.Context(), cart); err != nil {
		t.Fatalf("SetCart error: %v", err)
	}

	got, err := client.GetCart(t.Context(), "abc123")
	if err != nil {
		t.Fatalf("GetCart error: %v", err)
	}
	if got.Token != "abc123" {
		t.Errorf("got token %q want %q", got.Token, "abc123")
	}
	if len(got.LineItems) != 1 {
		t.Errorf("got %d line items want 1", len(got.LineItems))
	}
	if got.LineItems[0].Amount != 2500 {
		t.Errorf("got amount %d want 2500", got.LineItems[0].Amount)
	}
}

func TestKVGetCart_NotFound(t *testing.T) {
	srv, client := newKVServer(t)
	defer srv.Close()

	_, err := client.GetCart(t.Context(), "does-not-exist")
	if err != store.ErrCartNotFound {
		t.Errorf("expected ErrCartNotFound, got %v", err)
	}
}

func TestKVDeleteCart(t *testing.T) {
	srv, client := newKVServer(t)
	defer srv.Close()

	cart := &models.Cart{Token: "del123", LineItems: []models.LineItem{}}
	_ = client.SetCart(t.Context(), cart)

	if err := client.DeleteCart(t.Context(), "del123"); err != nil {
		t.Fatalf("DeleteCart error: %v", err)
	}

	_, err := client.GetCart(t.Context(), "del123")
	if err != store.ErrCartNotFound {
		t.Errorf("expected ErrCartNotFound after delete, got %v", err)
	}
}

func TestCartTotal(t *testing.T) {
	cart := &models.Cart{
		Token: "x",
		LineItems: []models.LineItem{
			{Amount: 1000, Quantity: 2},
			{Amount: 500, Quantity: 1},
		},
	}
	if cart.Total() != 2500 {
		t.Errorf("Total() = %d, want 2500", cart.Total())
	}
}

// Ensure JSON round-trip doesn't corrupt amounts (int64 boundary check)
func TestKVCartJSONRoundtrip(t *testing.T) {
	srv, client := newKVServer(t)
	defer srv.Close()

	cart := &models.Cart{
		Token: "rt",
		LineItems: []models.LineItem{
			{PriceID: "price_x", Amount: 999999999, Quantity: 1},
		},
	}
	_ = client.SetCart(t.Context(), cart)
	got, _ := client.GetCart(t.Context(), "rt")
	b1, _ := json.Marshal(cart)
	b2, _ := json.Marshal(got)
	if string(b1) != string(b2) {
		t.Errorf("JSON mismatch:\n  got  %s\n  want %s", b2, b1)
	}
}
