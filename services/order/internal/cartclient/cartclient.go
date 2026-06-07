// Package cartclient is the order service's client for the cart service. It
// demonstrates auth propagation: the order service forwards the customer's own
// bearer token so the cart service returns that customer's cart.
package cartclient

import (
	"context"
	"net/http"

	"github.com/aura-commerce/platform/pkg/restclient"
	"github.com/aura-commerce/platform/services/order/internal/domain"
)

// Client talks to the cart service over HTTP.
type Client struct {
	rc *restclient.Client
}

func New(baseURL string) *Client {
	return &Client{rc: restclient.New(baseURL)}
}

// cartResponse mirrors the cart service's GET /cart payload.
type cartResponse struct {
	Items []struct {
		ProductID  string `json:"productId"`
		Slug       string `json:"slug"`
		Name       string `json:"name"`
		PriceCents int    `json:"priceCents"`
		Qty        int    `json:"qty"`
	} `json:"items"`
}

// Get fetches the caller's cart and maps it onto order line items. A cart line
// and an order line have the same shape, so we return domain.OrderItem directly.
func (c *Client) Get(ctx context.Context, bearer string) ([]domain.OrderItem, error) {
	var resp cartResponse
	if err := c.rc.Get(ctx, "/api/v1/cart", bearer, &resp); err != nil {
		return nil, err
	}
	items := make([]domain.OrderItem, 0, len(resp.Items))
	for _, it := range resp.Items {
		items = append(items, domain.OrderItem{
			ProductID:  it.ProductID,
			Slug:       it.Slug,
			Name:       it.Name,
			PriceCents: it.PriceCents,
			Qty:        it.Qty,
		})
	}
	return items, nil
}

// Clear empties the caller's cart after a successful checkout.
func (c *Client) Clear(ctx context.Context, bearer string) error {
	return c.rc.Do(ctx, http.MethodDelete, "/api/v1/cart", bearer, nil, nil)
}
