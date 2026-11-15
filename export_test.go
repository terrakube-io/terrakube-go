package terrakube

import (
	"context"
	"net/http"
	"net/url"
)

// Unexported helpers exposed to the external test package.

// SameOrigin is sameOrigin.
var SameOrigin = sameOrigin

// RedirectAllowed is redirectAllowed.
func (c *Client) RedirectAllowed(target *url.URL) bool { return c.redirectAllowed(target) }

// RequestStream is requestStream.
func (c *Client) RequestStream(ctx context.Context, method, path string, hdr http.Header) (*http.Response, error) {
	return c.requestStream(ctx, method, path, hdr)
}

// RequestStreamURL is requestStreamURL.
func (c *Client) RequestStreamURL(ctx context.Context, method, rawURL string, hdr http.Header) (*http.Response, error) {
	return c.requestStreamURL(ctx, method, rawURL, hdr)
}
