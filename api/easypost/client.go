package easypost

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

const baseURL = "https://api.easypost.com/v2"

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

// Client makes EasyPost REST API calls.
type Client struct {
	apiKey   string
	fromAddr Address
	http     *http.Client
}

// NewClient constructs an EasyPost client with a fixed from-address.
func NewClient(apiKey string, from Address) *Client {
	return &Client{
		apiKey:   apiKey,
		fromAddr: from,
		http:     &http.Client{},
	}
}

// RateShop creates an EasyPost shipment and returns "shipmentID:rateID" for the
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
		Length float64 `json:"length"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
		Weight float64 `json:"weight"`
	}
	type shipmentBody struct {
		ToAddress   addrFields   `json:"to_address"`
		FromAddress addrFields   `json:"from_address"`
		Parcel      parcelFields `json:"parcel"`
	}

	reqBody, err := json.Marshal(map[string]any{
		"shipment": shipmentBody{
			ToAddress: addrFields{
				Name:    to.Name,
				Street1: to.Street1,
				Street2: to.Street2,
				City:    to.City,
				State:   to.State,
				Zip:     to.Zip,
				Country: to.Country,
			},
			FromAddress: addrFields{
				Name:    c.fromAddr.Name,
				Street1: c.fromAddr.Street1,
				City:    c.fromAddr.City,
				State:   c.fromAddr.State,
				Zip:     c.fromAddr.Zip,
				Country: c.fromAddr.Country,
			},
			// Fixed parcel profile: 12×9×1 in padded mailer, 8 oz
			Parcel: parcelFields{Length: 12, Width: 9, Height: 1, Weight: 8},
		},
	})
	if err != nil {
		return "", err
	}

	var result struct {
		ID    string `json:"id"`
		Rates []struct {
			ID   string `json:"id"`
			Rate string `json:"rate"`
		} `json:"rates"`
	}
	if err := c.post(ctx, "/shipments", reqBody, &result); err != nil {
		return "", fmt.Errorf("easypost shipment: %w", err)
	}

	var bestRateID string
	var bestAmount float64 = -1
	for _, r := range result.Rates {
		amt, err := strconv.ParseFloat(r.Rate, 64)
		if err != nil {
			continue
		}
		if bestRateID == "" || amt < bestAmount {
			bestRateID = r.ID
			bestAmount = amt
		}
	}
	if bestRateID == "" {
		return "", fmt.Errorf("easypost: no rates returned for shipment %s", result.ID)
	}
	return result.ID + ":" + bestRateID, nil
}

// BuyLabel purchases a shipping label. token is "shipmentID:rateID" from RateShop.
// Returns tracking number, carrier name, and label PDF URL.
func (c *Client) BuyLabel(ctx context.Context, token string) (trackingNumber, carrier, labelURL string, err error) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("easypost: invalid rate token %q", token)
	}
	shipmentID, rateID := parts[0], parts[1]

	body, err := json.Marshal(map[string]any{
		"rate": map[string]string{"id": rateID},
	})
	if err != nil {
		return "", "", "", err
	}

	var result struct {
		TrackingCode string `json:"tracking_code"`
		PostageLabel struct {
			LabelURL string `json:"label_url"`
		} `json:"postage_label"`
		SelectedRate struct {
			Carrier string `json:"carrier"`
		} `json:"selected_rate"`
	}
	if err := c.post(ctx, "/shipments/"+shipmentID+"/buy", body, &result); err != nil {
		return "", "", "", fmt.Errorf("easypost buy: %w", err)
	}
	if result.TrackingCode == "" || result.PostageLabel.LabelURL == "" {
		return "", "", "", fmt.Errorf("easypost: label purchase returned empty tracking or label URL")
	}
	return result.TrackingCode, result.SelectedRate.Carrier, result.PostageLabel.LabelURL, nil
}

func (c *Client) post(ctx context.Context, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.apiKey, "")
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
