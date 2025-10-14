package langfuse

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	Version = "v1.0.0"

	defaultBaseURL   = "https://cloud.langfuse.com/api/public/"
	defaultUserAgent = "go-client-langfuse" + "/" + Version
	defaultMediaType = "*/*"
)

var errNonNilContext = errors.New("context must be non-nil")

// Client represents a langfuse client with retryable HTTP capabilities
type Client struct {
	clientMu  sync.Mutex
	client    *http.Client
	BaseURL   *url.URL
	ApiToken  string
	UserAgent string

	common service // Reuse a single instance of service

	// Services
	Projects *ProjectsService
	Prompts  *PromptsService
}

// service is a base struct for different service types
type service struct {
	client *Client
}

// RequestOption represents an option that can modify an http.Request.
type RequestOption func(req *http.Request)

// roundTripperFunc creates a RoundTripper (transport).
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

// NewClient initializes and returns a new langfuse Client.
func (c *Client) NewRequest(method, urlStr string, body any, opts ...RequestOption) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("baseURL must have a trailing slash, but %q does not", c.BaseURL)
	}

	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", defaultMediaType)
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	for _, opt := range opts {
		opt(req)
	}

	return req, nil
}

// Do sends an HTTP request and decodes the response into v.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := c.BareDo(ctx, req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()
	return resp, err
}

// BareDo sends an HTTP request and returns an HTTP response.
func (c *Client) BareDo(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.bareDo(ctx, c.client, req)
}

// bareDo sends an HTTP request and returns an HTTP response, using the provided http.Client.
func (c *Client) bareDo(ctx context.Context, caller *http.Client, req *http.Request) (*http.Response, error) {
	if ctx == nil {
		return nil, errNonNilContext
	}

	resp, err := caller.Do(req)
	var response *http.Response
	if resp != nil {
		response = resp
	}
	if err != nil {
		return nil, err
	}
	return response, err
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	httpClient2 := *httpClient
	c := &Client{client: &httpClient2}
	c.initialize()
	return c
}

// WithAuthToken returns a copy of the client configured to use the provided token for the Authorization header.
func (c *Client) WithAuthToken(token string) *Client {
	c2 := c.copy()
	defer c2.initialize()
	transport := c2.client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	// Base64 encode the token to handle special characters
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	// Set the Authorization header for each request
	c2.client.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", encodedToken))
			return transport.RoundTrip(req)
		},
	)
	return c2
}

// initialize sets default values and initializes services.
func (c *Client) initialize() {
	if c.client == nil {
		c.client = &http.Client{}
	}
	if c.BaseURL == nil {
		c.BaseURL, _ = url.Parse(defaultBaseURL)
	}
	if c.UserAgent == "" {
		c.UserAgent = defaultUserAgent
	}
	c.common.client = c
	c.Projects = (*ProjectsService)(&c.common)
	c.Prompts = (*PromptsService)(&c.common)
}

// copy returns a copy of the current client. It must be initialized before use.
func (c *Client) copy() *Client {
	c.clientMu.Lock()
	// can't use *c here because that would copy mutexes by value.
	clone := Client{
		client:    &http.Client{},
		UserAgent: c.UserAgent,
		BaseURL:   c.BaseURL,
	}
	c.clientMu.Unlock()
	if c.client != nil {
		clone.client.Transport = c.client.Transport
		clone.client.CheckRedirect = c.client.CheckRedirect
		clone.client.Jar = c.client.Jar
		clone.client.Timeout = c.client.Timeout
	}
	return &clone
}
