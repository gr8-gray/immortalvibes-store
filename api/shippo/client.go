package shippo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const baseURL = "https://api.goshippo.com"

// Address is a postal address used for both from and to.
type Address struct {
	Name    string
	Street1 string
	Street2 string
	City    string
	State   string
	Zip     string
	Country string
	Email   string // Shippo requires a non-empty email on address_from to purchase a label.
	Phone   string // USPS requires a non-empty phone on address_from to purchase a label.
}

// Client makes Shippo REST API calls.
type Client struct {
	apiKey   string
	fromAddr Address
	http     *http.Client
}

// NewClient constructs a Shippo client with a fixed from-address.
func NewClient(apiKey string, from Address) *Client {
	return &Client{
		apiKey:   apiKey,
		fromAddr: from,
		http:     &http.Client{},
	}
}

// RateShop creates a Shippo shipment and returns "rateID:carrier" for the
// cheapest available rate. The opaque token is consumed by BuyLabel.
func (c *Client) RateShop(ctx context.Context, to Address) (string, error) {
	type addrFields struct {
		Name    string `json:"name"`
		Street1 string `json:"street1"`
		Street2 string `json:"street2,omitempty"`
		City    string `json:"city"`
		State   string `json:"state"`
		Zip     string `json:"zip"`
		Country string `json:"country"`
		Email   string `json:"email,omitempty"`
		Phone   string `json:"phone,omitempty"`
	}
	type parcelFields struct {
		Length       float64 `json:"length"`
		Width        float64 `json:"width"`
		Height       float64 `json:"height"`
		DistanceUnit string  `json:"distance_unit"`
		Weight       float64 `json:"weight"`
		MassUnit     string  `json:"mass_unit"`
	}

	body, err := json.Marshal(map[string]any{
		"address_from": addrFields{
			Name:    c.fromAddr.Name,
			Street1: c.fromAddr.Street1,
			City:    c.fromAddr.City,
			State:   c.fromAddr.State,
			Zip:     c.fromAddr.Zip,
			Country: c.fromAddr.Country,
			Email:   c.fromAddr.Email,
			Phone:   c.fromAddr.Phone,
		},
		"address_to": addrFields{
			Name:    to.Name,
			Street1: to.Street1,
			Street2: to.Street2,
			City:    to.City,
			State:   to.State,
			Zip:     to.Zip,
			Country: to.Country,
			Email:   to.Email,
			Phone:   to.Phone,
		},
		// Fixed parcel profile: 12×9×1 in padded mailer, 8 oz
		"parcels": []parcelFields{{
			Length: 12, Width: 9, Height: 1, DistanceUnit: "in",
			Weight: 8, MassUnit: "oz",
		}},
		"async": false,
	})
	if err != nil {
		return "", err
	}

	var result struct {
		ObjectID string `json:"object_id"`
		Rates    []struct {
			ObjectID string `json:"object_id"`
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
			Provider string `json:"provider"`
		} `json:"rates"`
	}
	if err := c.do(ctx, "/shipments/", body, &result); err != nil {
		return "", fmt.Errorf("shippo shipment: %w", err)
	}

	var bestRateID, bestCarrier, bestCurrency string
	var bestAmount float64 = -1
	for _, r := range result.Rates {
		amt, err := strconv.ParseFloat(r.Amount, 64)
		if err != nil {
			continue
		}
		if bestRateID == "" || amt < bestAmount {
			bestRateID = r.ObjectID
			bestCarrier = r.Provider
			bestCurrency = r.Currency
			bestAmount = amt
		}
	}
	if bestRateID == "" {
		return "", fmt.Errorf("shippo: no rates returned for shipment %s", result.ObjectID)
	}
	// Token packs rateID:carrier:amount:currency so BuyLabel can report the
	// label cost to the owner without a second Shippo round-trip.
	return strings.Join([]string{
		bestRateID, bestCarrier, strconv.FormatFloat(bestAmount, 'f', 2, 64), bestCurrency,
	}, ":"), nil
}

// RateEstimate holds pricing info for the cheapest rate returned by Shippo.
type RateEstimate struct {
	Provider string  // e.g. "USPS"
	Service  string  // e.g. "Priority Mail"
	Amount   float64 // dollars
	Currency string  // e.g. "USD"
}

// EstimateRate creates a Shippo shipment and returns the cheapest rate details.
// Unlike RateShop, it returns full rate info suitable for display pre-purchase.
func (c *Client) EstimateRate(ctx context.Context, to Address) (*RateEstimate, error) {
	type addrFields struct {
		Name    string `json:"name"`
		Street1 string `json:"street1"`
		Street2 string `json:"street2,omitempty"`
		City    string `json:"city"`
		State   string `json:"state"`
		Zip     string `json:"zip"`
		Country string `json:"country"`
		Email   string `json:"email,omitempty"`
		Phone   string `json:"phone,omitempty"`
	}
	type parcelFields struct {
		Length       float64 `json:"length"`
		Width        float64 `json:"width"`
		Height       float64 `json:"height"`
		DistanceUnit string  `json:"distance_unit"`
		Weight       float64 `json:"weight"`
		MassUnit     string  `json:"mass_unit"`
	}

	body, err := json.Marshal(map[string]any{
		"address_from": addrFields{
			Name:    c.fromAddr.Name,
			Street1: c.fromAddr.Street1,
			City:    c.fromAddr.City,
			State:   c.fromAddr.State,
			Zip:     c.fromAddr.Zip,
			Country: c.fromAddr.Country,
			Email:   c.fromAddr.Email,
			Phone:   c.fromAddr.Phone,
		},
		"address_to": addrFields{
			Name:    to.Name,
			Street1: to.Street1,
			Street2: to.Street2,
			City:    to.City,
			State:   to.State,
			Zip:     to.Zip,
			Country: to.Country,
			Email:   to.Email,
			Phone:   to.Phone,
		},
		"parcels": []parcelFields{{
			Length: 12, Width: 9, Height: 1, DistanceUnit: "in",
			Weight: 8, MassUnit: "oz",
		}},
		"async": false,
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Rates []struct {
			Amount            string `json:"amount"`
			Currency          string `json:"currency"`
			Provider          string `json:"provider"`
			ServicelevelToken string `json:"servicelevel_token"`
			// Shippo wraps servicelevel in a nested object with {name, token}.
			// attributes is returned as []string (e.g. "BESTVALUE", "CHEAPEST") and unused here.
			Servicelevel *struct {
				Name  string `json:"name"`
				Token string `json:"token"`
			} `json:"servicelevel"`
		} `json:"rates"`
	}
	if err := c.do(ctx, "/shipments/", body, &result); err != nil {
		return nil, fmt.Errorf("shippo estimate: %w", err)
	}

	var best *RateEstimate
	var bestAmount float64 = -1
	for _, r := range result.Rates {
		amt, err := strconv.ParseFloat(r.Amount, 64)
		if err != nil {
			continue
		}
		if best == nil || amt < bestAmount {
			var service string
			if r.Servicelevel != nil {
				service = r.Servicelevel.Name
			}
			best = &RateEstimate{
				Provider: r.Provider,
				Service:  service,
				Amount:   amt,
				Currency: r.Currency,
			}
			bestAmount = amt
		}
	}
	if best == nil {
		return nil, fmt.Errorf("shippo: no rates returned")
	}
	return best, nil
}

// BuyLabel purchases a shipping label. token is "rateID:carrier:amount:currency"
// from RateShop. Returns tracking number, carrier name, label PDF URL, and the
// label cost (e.g. "6.74 USD") for owner-facing reporting.
func (c *Client) BuyLabel(ctx context.Context, token string) (trackingNumber, carrier, labelURL, cost string, err error) {
	parts := strings.SplitN(token, ":", 4)
	if len(parts) < 2 {
		return "", "", "", "", fmt.Errorf("shippo: invalid rate token %q", token)
	}
	rateID, carrierName := parts[0], parts[1]
	if len(parts) == 4 && parts[2] != "" {
		cost = parts[2] + " " + strings.ToUpper(parts[3])
	}

	body, err := json.Marshal(map[string]any{
		"rate":            rateID,
		"label_file_type": "PDF",
		"async":           false,
	})
	if err != nil {
		return "", "", "", "", err
	}

	var result struct {
		TrackingNumber string `json:"tracking_number"`
		LabelURL       string `json:"label_url"`
		ObjectState    string `json:"object_state"`
	}
	if err := c.do(ctx, "/transactions/", body, &result); err != nil {
		return "", "", "", "", fmt.Errorf("shippo buy: %w", err)
	}
	if result.TrackingNumber == "" || result.LabelURL == "" {
		return "", "", "", "", fmt.Errorf("shippo: label purchase returned empty tracking or label URL (state: %s)", result.ObjectState)
	}
	return result.TrackingNumber, carrierName, result.LabelURL, cost, nil
}

func (c *Client) do(ctx context.Context, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "ShippoToken "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, respBody)
	}
	return json.Unmarshal(respBody, out)
}
