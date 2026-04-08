package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/immortalvibes/api/models"
)

// ErrCartNotFound is returned when the cart key does not exist in KV.
var ErrCartNotFound = errors.New("cart not found")

// KVClient wraps the Cloudflare KV REST API for cart storage.
type KVClient struct {
	baseURL     string // e.g. "https://api.cloudflare.com"
	accountID   string
	namespaceID string
	apiToken    string
	http        *http.Client
}

// NewKVClient constructs a KVClient. baseURL is overridable for tests.
func NewKVClient(baseURL, accountID, namespaceID, apiToken string) *KVClient {
	return &KVClient{
		baseURL:     baseURL,
		accountID:   accountID,
		namespaceID: namespaceID,
		apiToken:    apiToken,
		http:        &http.Client{},
	}
}

func (c *KVClient) kvURL(key string) string {
	return fmt.Sprintf(
		"%s/client/v4/accounts/%s/storage/kv/namespaces/%s/values/%s",
		c.baseURL, c.accountID, c.namespaceID, key,
	)
}

// GetCart retrieves and deserialises the cart for the given token.
// Returns ErrCartNotFound if the key does not exist.
func (c *KVClient) GetCart(ctx context.Context, token string) (*models.Cart, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.kvURL(token), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrCartNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("KV GET %s: status %d: %s", token, resp.StatusCode, body)
	}

	var cart models.Cart
	if err := json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		return nil, fmt.Errorf("KV decode: %w", err)
	}
	return &cart, nil
}

// SetCart serialises the cart and writes it to KV under cart.Token.
func (c *KVClient) SetCart(ctx context.Context, cart *models.Cart) error {
	body, err := json.Marshal(cart)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.kvURL(cart.Token), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("KV PUT %s: status %d: %s", cart.Token, resp.StatusCode, b)
	}
	return nil
}

// DeleteCart removes the cart key from KV. Idempotent — not-found is not an error.
func (c *KVClient) DeleteCart(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.kvURL(token), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("KV DELETE %s: status %d: %s", token, resp.StatusCode, b)
	}
	return nil
}
