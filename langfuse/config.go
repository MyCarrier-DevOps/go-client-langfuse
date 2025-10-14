package langfuse

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/viper"
)

// Config holds the configuration for ArgoCD client operations.
// Only ServerUrl and AuthToken are required fields; AppName and Revision are optional.
type Config struct {
	ServerUrl string `mapstructure:"server_url"`
	ApiToken  string `mapstructure:"api_token"`
}

// config is an alias for Config to avoid import cycles in other packages
var config Config

// ConfigLoaderInterface defines the contract for configuration loading
type ConfigLoaderInterface interface {
	// LoadConfig loads configuration from environment variables
	LoadConfig() (*Config, error)
	// ValidateConfig validates the provided configuration
	ValidateConfig(config *Config) error
}

// LoadConfig loads the Langfuse client configuration from environment variables.
// It binds the following environment variables to configuration fields:
//
//   - LANGFUSE_SERVER_URL -> ServerUrl (required)
//   - LANGFUSE_API_TOKEN -> ApiToken (required)
//
// Returns an error if required environment variables are missing or if
// there are issues with configuration binding or validation.
func LoadConfig() error {
	if err := viper.BindEnv("server_url", "LANGFUSE_SERVER_URL"); err != nil {
		return fmt.Errorf("error binding LANGFUSE_SERVER_URL: %w", err)
	}
	if err := viper.BindEnv("api_token", "LANGFUSE_API_TOKEN"); err != nil {
		return fmt.Errorf("error binding LANGFUSE_API_TOKEN: %w", err)
	}

	viper.AutomaticEnv()

	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return fmt.Errorf("error validating config: %w", err)
	}

	return nil
}

// validateConfig validates that all required configuration fields are present.
// Currently only validates ServerUrl and ApiToken as required fields.
func validateConfig(config *Config) error {
	if config.ServerUrl == "" {
		return fmt.Errorf("LANGFUSE_SERVER_URL is required")
	}

	if config.ApiToken == "" {
		return fmt.Errorf("LANGFUSE_API_TOKEN is required")
	} else {
		// convert to base64 if not already
		config.ApiToken = base64.StdEncoding.EncodeToString([]byte(config.ApiToken))
	}

	return nil
}
