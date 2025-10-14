package langfuse

import (
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Config represents the configuration for langfuse client operations.
type Config struct {
	ServerUrl string `mapstructure:"server_url"`
	ApiToken  string `mapstructure:"api_token"`
}

// Client represents a langfuse client with retryable HTTP capabilities
type Client struct {
	retryableClient *retryablehttp.Client
	baseUrl         string
	apiToken        string
}

// NewClient creates a new langfuse client with retryable HTTP configuration
func NewClient(config *Config) *Client {
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

// service is a base struct for different service types
type service struct {
	client *Client
}
