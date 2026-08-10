package bachs

import "net/http"

// Option configures a *Client created by NewClient.
type Option func(*Client)

// WithBaseURL overrides the base URL the client sends requests to. The URL
// must not include a trailing slash or the /v1 path segment; both are handled
// by the client.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithSandbox explicitly targets the sandbox environment
// (https://sandbox-api.bachs.io). This is the default; the option exists so
// callers can state the environment without relying on defaults.
func WithSandbox() Option {
	return func(c *Client) {
		c.baseURL = BaseURLSandbox
	}
}

// WithProduction targets the production environment
// (https://api.bachs.io). Never the default: production is only ever used when
// explicitly requested.
func WithProduction() Option {
	return func(c *Client) {
		c.baseURL = BaseURLProduction
	}
}

// WithHTTPClient replaces the client's HTTP transport. The provided client is
// used as-is, including its Timeout and Transport. A nil argument is ignored.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}
