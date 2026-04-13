package shippo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

// RateShop creates a shipment with the fixed parcel profile and returns the
// object_id of the cheapest valid rate.
func (c *Client) RateShop(ctx context.Context, to Address) (string, error) {
	body, err := json.Marshal(map[string]any{
		"address_from": map[string]string{
			"name":    c.fromAddr.Name,
			"street1": c.fromAddr.Street1,
			"city":    c.fromAddr.City,
			"state":   c.fromAddr.State,
			"zip":     c.fromAddr.Zip,
			"country": c.fromAddr.Country,
		},
		"address_to": map[string]string{
			"name":    to.Name,
			"street1": to.Street1,
			"street2": to.Street2,
			"city":    to.City,
			"state":   to.State,
			"zip":     to.Zip,
			"country": to.Country,
		},
		"parcels": []map[string]string{{
			"length":        "12",
			"width":         "9",
			"height":        "1",
			"distance_unit": "in",
			"weight":        "8",
			"mass_unit":     "oz",
		}},
		"async": false,
	})
	if err != nil {
		return "", err
	}

	var result struct {
		Rates []struct {
			ObjectID     string `json:"object_id"`
			ObjectStatus string `json:"object_status"`
			Amount       string `json:"amount"`
		} `json:"rates"`
	}
	if err := c.post(ctx, "/shipments/", body, &result); err != nil {
		return "", fmt.Errorf("shippo shipment: %w", err)
	}

	var bestID string
	var bestAmount float64 = -1
	for _, r := range result.Rates {
		if r.ObjectStatus != "VALID" {
			continue
		}
		amt, err := strconv.ParseFloat(r.Amount, 64)
		if err != nil {
			continue
		}
		if bestID == "" || amt < bestAmount {
			bestID = r.ObjectID
			bestAmount = amt
		}
	}
	if bestID == "" {
		return "", fmt.Errorf("shippo: no valid rates returned")
	}
	return bestID, nil
}

// BuyLabel purchases a PDF label for the given rate ID.
// Returns tracking number, carrier name, and label PDF URL.
func (c *Client) BuyLabel(ctx context.Context, rateID string) (trackingNumber, carrier, labelURL string, err error) {
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
		Rate           struct {
			Provider string `json:"provider"`
		} `json:"rate"`
	}
	if err := c.post(ctx, "/transactions/", body, &result); err != nil {
		return "", "", "", fmt.Errorf("shippo transaction: %w", err)
	}
	if result.TrackingNumber == "" || result.LabelURL == "" {
		return "", "", "", fmt.Errorf("shippo: label purchase returned empty tracking or label URL")
	}
	return result.TrackingNumber, result.Rate.Provider, result.LabelURL, nil
}

func (c *Client) post(ctx context.Context, path string, body []byte, out any) error {
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
