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
		},
		"address_to": addrFields{
			Name:    to.Name,
			Street1: to.Street1,
			Street2: to.Street2,
			City:    to.City,
			State:   to.State,
			Zip:     to.Zip,
			Country: to.Country,
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
			Provider string `json:"provider"`
		} `json:"rates"`
	}
	if err := c.do(ctx, "/shipments/", body, &result); err != nil {
		return "", fmt.Errorf("shippo shipment: %w", err)
	}

	var bestRateID, bestCarrier string
	var bestAmount float64 = -1
	for _, r := range result.Rates {
		amt, err := strconv.ParseFloat(r.Amount, 64)
		if err != nil {
			continue
		}
		if bestRateID == "" || amt < bestAmount {
			bestRateID = r.ObjectID
			bestCarrier = r.Provider
			bestAmount = amt
		}
	}
	if bestRateID == "" {
		return "", fmt.Errorf("shippo: no rates returned for shipment %s", result.ObjectID)
	}
	return bestRateID + ":" + bestCarrier, nil
}

// BuyLabel purchases a shipping label. token is "rateID:carrier" from RateShop.
// Returns tracking number, carrier name, and label PDF URL.
func (c *Client) BuyLabel(ctx context.Context, token string) (trackingNumber, carrier, labelURL string, err error) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("shippo: invalid rate token %q", token)
	}
	rateID, carrierName := parts[0], parts[1]

	body, err := json.Marshal(map[string]any{
		"rate":            rateID,
		"label_file_type": "PDF",
		"async":           false,
	})
	if err != nil {
		return "", "", "", err
	}

	var result struct {
		TrackingNumber string `json:"tracking_number"`
		LabelURL       string `json:"label_url"`
		ObjectState    string `json:"object_state"`
	}
	if err := c.do(ctx, "/transactions/", body, &result); err != nil {
		return "", "", "", fmt.Errorf("shippo buy: %w", err)
	}
	if result.TrackingNumber == "" || result.LabelURL == "" {
		return "", "", "", fmt.Errorf("shippo: label purchase returned empty tracking or label URL (state: %s)", result.ObjectState)
	}
	return result.TrackingNumber, carrierName, result.LabelURL, nil
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
