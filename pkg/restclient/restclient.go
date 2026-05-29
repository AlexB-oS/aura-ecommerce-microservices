// Package restclient is a tiny HTTP client for synchronous service-to-service
// calls inside the platform (e.g. cart → product, order → cart). It understands
// the platform's JSON envelope ({ "data": … } / { "error": … }) so callers work
// with plain Go structs and a typed error.
package restclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aura-commerce/platform/pkg/logger"
)

// Error carries the HTTP status and the platform error code from a failed call,
// so callers can react (e.g. treat 404 as "product not found").
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("rest call failed: %d %s: %s", e.Status, e.Code, e.Message)
}

// Client targets one upstream service base URL (e.g. http://product-service:8080).
type Client struct {
	base string
	http *http.Client
}

// New builds a client with a sensible timeout for internal calls.
func New(base string) *Client {
	return &Client{base: base, http: &http.Client{Timeout: 5 * time.Second}}
}

// Do performs the request and unmarshals the response envelope's `data` into
// out (which may be nil if no body is expected). `bearer`, if non-empty, is
// forwarded as the Authorization header so the caller's identity propagates.
// The inbound request id is forwarded too, keeping a call traceable end-to-end.
func (c *Client) Do(ctx context.Context, method, path, bearer string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if rid := logger.RequestID(ctx); rid != "" {
		req.Header.Set("X-Request-ID", rid)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var env struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(payload, &env)
		return &Error{Status: resp.StatusCode, Code: env.Error.Code, Message: env.Error.Message}
	}

	if out == nil {
		return nil
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(payload, &env); err != nil {
		return err
	}
	return json.Unmarshal(env.Data, out)
}

// Get is a convenience wrapper for GET.
func (c *Client) Get(ctx context.Context, path, bearer string, out any) error {
	return c.Do(ctx, http.MethodGet, path, bearer, nil, out)
}

// StatusOf returns the HTTP status from an error returned by this package, or 0.
func StatusOf(err error) int {
	if e, ok := err.(*Error); ok {
		return e.Status
	}
	return 0
}
