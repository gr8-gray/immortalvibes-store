package handlers

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strings"

	"github.com/immortalvibes/api/shippo"
)

// ShippoEstimator is the subset of shippo.Client used by ShippingHandler.
type ShippoEstimator interface {
	EstimateRate(ctx context.Context, to shippo.Address) (*shippo.RateEstimate, error)
}

// ShippingHandler handles POST /api/shipping/estimate.
type ShippingHandler struct {
	shippo   ShippoEstimator
	fromAddr shippo.Address
}

// NewShippingHandler constructs a ShippingHandler.
// fromAddr fields are optional; empty fields fall back to hardcoded defaults.
func NewShippingHandler(client ShippoEstimator, fromAddr shippo.Address) *ShippingHandler {
	if fromAddr.Name == "" {
		fromAddr.Name = "Immortal Vibes"
	}
	if fromAddr.Street1 == "" {
		fromAddr.Street1 = "123 Main St"
	}
	if fromAddr.City == "" {
		fromAddr.City = "Los Angeles"
	}
	if fromAddr.State == "" {
		fromAddr.State = "CA"
	}
	if fromAddr.Zip == "" {
		fromAddr.Zip = "90001"
	}
	if fromAddr.Country == "" {
		fromAddr.Country = "US"
	}
	return &ShippingHandler{shippo: client, fromAddr: fromAddr}
}

type shippingEstimateRequest struct {
	Name    string `json:"name"`
	Street1 string `json:"street1"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

type shippingRate struct {
	Provider string `json:"provider"`
	Service  string `json:"service"`
	Amount   int    `json:"amount"`   // cents
	Currency string `json:"currency"` // uppercase ISO, e.g. "USD"
}

type shippingEstimateResponse struct {
	Rate  *shippingRate `json:"rate"`
	Error string        `json:"error,omitempty"`
}

// Estimate handles POST /api/shipping/estimate.
func (h *ShippingHandler) Estimate(w http.ResponseWriter, r *http.Request) {
	var req shippingEstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(shippingEstimateResponse{Error: "invalid request body"})
		return
	}

	if req.Street1 == "" || req.City == "" || req.State == "" || req.Zip == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(shippingEstimateResponse{Error: "street1, city, state, and zip are required"})
		return
	}

	country := req.Country
	if country == "" {
		country = "US"
	}

	to := shippo.Address{
		Name:    req.Name,
		Street1: req.Street1,
		City:    req.City,
		State:   req.State,
		Zip:     req.Zip,
		Country: country,
	}

	estimate, err := h.shippo.EstimateRate(r.Context(), to)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusOK) // return 200 with null rate per spec
		json.NewEncoder(w).Encode(shippingEstimateResponse{Error: "Unable to calculate shipping"})
		return
	}

	cents := int(math.Round(estimate.Amount * 100))
	json.NewEncoder(w).Encode(shippingEstimateResponse{
		Rate: &shippingRate{
			Provider: estimate.Provider,
			Service:  estimate.Service,
			Amount:   cents,
			Currency: strings.ToUpper(estimate.Currency),
		},
	})
}
