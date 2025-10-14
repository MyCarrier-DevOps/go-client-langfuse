package langfuse

import (
	"fmt"
	"io"
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
	apiToken        string

	Projects *ProjectsService
	Prompts  *PromptsService
}

type service struct {
	client *Client
}

// NewClient creates a new ArgoCD client with retryable HTTP configuration
func (config *Config) NewClient() *Client {
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

	return &Client{
		retryableClient: retryClient,
		baseUrl:         config.ServerUrl,
		apiToken:        config.ApiToken,
	}
}

func (c *Client) Do(uri string) (body []byte, err error) {
	req, err := retryablehttp.NewRequest("GET", c.baseUrl+uri, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if c.apiToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")

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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("client error %d: %s", resp.StatusCode, string(body))
	}

	// Handle any remaining 5xx errors that exhausted retries
	if resp.StatusCode >= 500 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	return body, nil
}
