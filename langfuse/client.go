package langfuse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	Version = "v1.0.0"

	defaultBaseURL   = "https://cloud.langfuse.com/api/public/"
	defaultUserAgent = "go-langfuse-client" + "/" + Version
	defaultMediaType = "*/*"
)

// Client represents an ArgoCD client with retryable HTTP capabilities
type Client struct {
	retryableClient *retryablehttp.Client
	baseUrl         string
	base64Token     string

	Projects *ProjectsService
	Prompts  *PromptsService
}

type service struct {
	client *Client
}

// NewClient creates a new Langfuse client with retryable HTTP configuration
// using the provided Config. This allows creating a client without relying on
// environment variables or the global config.
//
// Example:
//
//	config, err := langfuse.NewConfig(
//	    "https://cloud.langfuse.com",
//	    "pk-lf-xxx",
//	    "sk-lf-xxx",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	client := langfuse.NewClient(config)
func NewClient(cfg *Config) *Client {
	retryClient := retryablehttp.NewClient()

	// Configure retry parameters
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 4 * time.Second
	retryClient.Backoff = retryablehttp.DefaultBackoff

	// Use default retry policy (retries on 5xx and network errors)
	retryClient.CheckRetry = retryablehttp.DefaultRetryPolicy

	// Disable default logging to avoid noise
	retryClient.Logger = nil

	client := &Client{
		retryableClient: retryClient,
		baseUrl:         cfg.ServerUrl,
		base64Token:     cfg.Base64Token,
	}

	// Initialize services with client reference
	client.Projects = (*ProjectsService)(&service{client: client})
	client.Prompts = (*PromptsService)(&service{client: client})

	return client
}

func (c *Client) Do(method, uri string) (body []byte, err error) {
	return c.DoWithBody(method, uri, nil)
}

func (c *Client) DoWithBody(method, uri string, payload interface{}) (body []byte, err error) {
	if method == "" {
		method = "GET"
	}

	var reqBody io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("error marshalling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// Parse the full URL to properly handle percent-encoded path segments
	fullURL := c.baseUrl + uri
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	// Set RawPath to preserve percent-encoding in the path
	// This is necessary because url.Parse decodes the path by default
	parsedURL.RawPath = parsedURL.EscapedPath()

	req, err := retryablehttp.NewRequest(method, parsedURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Manually set the URL with RawPath preserved
	req.URL = parsedURL

	if c.base64Token == "" {
		return nil, fmt.Errorf("Base64 token is required")
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.base64Token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", defaultMediaType)
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := c.retryableClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return
		}
	}()

	// Handle 4xx client errors (these weren't retried)
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		return nil, fmt.Errorf("client error %d: %s", resp.StatusCode, string(body))
	}

	// Handle any remaining 5xx errors that exhausted retries
	if resp.StatusCode >= 500 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	return body, nil
}
