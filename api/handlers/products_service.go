package handlers

import (
	"context"
	"fmt"

	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/store"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
)

// StripeProductService is the live implementation of ProductService.
// It lists active Stripe products + prices, enriches with R2 URLs and DB stock.
type StripeProductService struct {
	stripeKey string
	r2BaseURL string
	db        *store.DB
}

// NewStripeProductService constructs the live service.
func NewStripeProductService(stripeKey, r2BaseURL string, db *store.DB) *StripeProductService {
	stripe.Key = stripeKey
	return &StripeProductService{
		stripeKey: stripeKey,
		r2BaseURL: r2BaseURL,
		db:        db,
	}
}

// ListProducts fetches all active Stripe products with their prices.
func (s *StripeProductService) ListProducts(ctx context.Context) ([]models.Product, error) {
	params := &stripe.ProductListParams{}
	params.Filters.AddFilter("active", "", "true")
	iter := product.List(params)

	var products []models.Product
	for iter.Next() {
		p := iter.Product()
		prod, err := s.enrichProduct(ctx, p)
		if err != nil {
			return nil, err
		}
		products = append(products, *prod)
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("stripe product list: %w", err)
	}
	return products, nil
}

// GetProduct fetches a single Stripe product by ID.
func (s *StripeProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	p, err := product.Get(id, nil)
	if err != nil {
		return nil, ErrProductNotFound
	}
	return s.enrichProduct(ctx, p)
}

func (s *StripeProductService) enrichProduct(ctx context.Context, p *stripe.Product) (*models.Product, error) {
	prices, err := s.fetchPrices(p.ID)
	if err != nil {
		return nil, err
	}

	imageURL := ""
	if len(p.Images) > 0 {
		imageURL = p.Images[0]
	}

	stock, err := s.db.GetStock(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("get stock for %s: %w", p.ID, err)
	}

	return &models.Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		ImageURL:    imageURL,
		Prices:      prices,
		StockCount:  stock,
	}, nil
}

func (s *StripeProductService) fetchPrices(productID string) ([]models.Price, error) {
	params := &stripe.PriceListParams{}
	params.Filters.AddFilter("product", "", productID)
	params.Filters.AddFilter("active", "", "true")
	iter := price.List(params)

	var prices []models.Price
	for iter.Next() {
		p := iter.Price()
		prices = append(prices, models.Price{
			PriceID:  p.ID,
			Currency: string(p.Currency),
			Amount:   p.UnitAmount,
		})
	}
	return prices, iter.Err()
}
