package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/immortalvibes/api/handlers"
)

func validatePromo(t *testing.T, d *fakeDiscounts, code string) (int, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"code": code})
	w := httptest.NewRecorder()
	handlers.NewPromoHandler(d).Validate(w, httptest.NewRequest(http.MethodPost, "/api/promo/validate", bytes.NewReader(b)))
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return w.Code, body
}

func TestPromoValidate(t *testing.T) {
	d := newFakeDiscounts()
	d.codes["FIVE"] = handlers.Discount{Code: "FIVE", CouponID: "cpn_5", AmountOff: 500, Currency: "usd", ProductIDs: []string{"prod_hat"}, MinimumAmount: 2000}

	code, body := validatePromo(t, d, "vibe10")
	disc, _ := body["discount"].(map[string]any)
	if code != http.StatusOK || body["valid"] != true || disc["type"] != "percent_off" || disc["value"] != 10.0 {
		t.Errorf("VIBE10: %d %v", code, body)
	}

	_, body = validatePromo(t, d, "FIVE")
	disc, _ = body["discount"].(map[string]any)
	if disc["type"] != "amount_off" || disc["value"] != 500.0 || disc["minimum_amount"] != 2000.0 || len(disc["product_ids"].([]any)) != 1 {
		t.Errorf("FIVE: %v", body)
	}

	if _, body = validatePromo(t, d, "NOPE"); body["valid"] != false || body["error"] != "Invalid code" {
		t.Errorf("unknown: %v", body)
	}

	d.err = handlers.ErrDiscountExhausted
	if _, body = validatePromo(t, d, "VIBE10"); body["valid"] != false || body["error"] != "This code is no longer available" {
		t.Errorf("exhausted: %v", body)
	}

	d.err = errors.New("stripe down")
	if code, body = validatePromo(t, d, "VIBE10"); code != http.StatusBadGateway || body["valid"] != false {
		t.Errorf("upstream failure: %d %v", code, body)
	}
}
