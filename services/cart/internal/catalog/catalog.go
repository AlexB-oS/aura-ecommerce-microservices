// Package catalog is the cart service's client for the product service. It is
// the concrete implementation of the service.Catalog port, demonstrating
// synchronous (REST) service-to-service communication: when a user adds an item
// the cart asks the product service for the authoritative name, price and
// stock.
package catalog

import (
	"context"

	"github.com/aura-commerce/platform/pkg/restclient"
	"github.com/aura-commerce/platform/services/cart/internal/domain"
)

// Client talks to the product service over HTTP.
type Client struct {
	rc *restclient.Client
}

// New builds a catalog client for the given product-service base URL.
func New(baseURL string) *Client {
	return &Client{rc: restclient.New(baseURL)}
}

// productResponse mirrors the product service's GET /products/{slug} payload
// (only the fields the cart cares about).
type productResponse struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
	Stock      int    `json:"stock"`
}

// Lookup fetches a product by slug. A 404 from the product service is mapped to
// domain.ErrProductUnavailable so the service layer stays transport-agnostic.
func (c *Client) Lookup(ctx context.Context, slug string) (*domain.ProductInfo, error) {
	var resp productResponse
	err := c.rc.Get(ctx, "/api/v1/products/"+slug, "", &resp)
	if err != nil {
		if restclient.StatusOf(err) == 404 {
			return nil, domain.ErrProductUnavailable
		}
		return nil, err
	}
	return &domain.ProductInfo{
		ID:         resp.ID,
		Slug:       resp.Slug,
		Name:       resp.Name,
		PriceCents: resp.PriceCents,
		Stock:      resp.Stock,
	}, nil
}
