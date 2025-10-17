package langfuse

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/viper"
)

// Config holds the configuration for ArgoCD client operations.
// Only ServerUrl and AuthToken are required fields; AppName and Revision are optional.
type Config struct {
	ServerUrl   string `mapstructure:"server_url"`
	PublicKey   string `mapstructure:"public_key"`
	SecretKey   string `mapstructure:"secret_key"`
	Base64Token string `mapstructure:"base64_token"`
}

// config is an alias for Config to avoid import cycles in other packages
var config Config

// NewConfig creates a new Config instance with the provided values.
// This allows creating a configuration without relying on environment variables.
// All three parameters are required.
//
// Example:
//
//	config := langfuse.NewConfig(
//	    "https://cloud.langfuse.com",
//	    "pk-lf-xxx",
//	    "sk-lf-xxx",
//	)
func NewConfig(serverUrl, publicKey, secretKey string) (*Config, error) {
	cfg := &Config{
		ServerUrl: serverUrl,
		PublicKey: publicKey,
		SecretKey: secretKey,
	}

	if cfg.PublicKey != "" && cfg.SecretKey != "" {
		cfg.Base64Token = base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", cfg.PublicKey, cfg.SecretKey)))
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ConfigLoaderInterface defines the contract for configuration loading
type ConfigLoaderInterface interface {
	// LoadConfigFromEnvVars loads configuration from environment variables
	LoadConfigFromEnvVars() (*Config, error)
	// ValidateConfig validates the provided configuration
	ValidateConfig(config *Config) error
}

// LoadConfigFromEnvVars loads the Langfuse client configuration from environment variables.
// This is an optional way to configure the client. Alternatively, use NewConfig()
// to create a configuration directly without environment variables.
//
// It binds the following environment variables to configuration fields:
//
//   - LANGFUSE_SERVER_URL -> ServerUrl (required)
//   - LANGFUSE_PUBLIC_KEY -> PublicKey (required)
//   - LANGFUSE_SECRET_KEY -> SecretKey (required)
//
// Returns an error if required environment variables are missing or if
// there are issues with configuration binding or validation.
func LoadConfigFromEnvVars() error {
	if err := viper.BindEnv("server_url", "LANGFUSE_SERVER_URL"); err != nil {
		return fmt.Errorf("error binding LANGFUSE_SERVER_URL: %w", err)
	}
	if err := viper.BindEnv("public_key", "LANGFUSE_PUBLIC_KEY"); err != nil {
		return fmt.Errorf("error binding LANGFUSE_PUBLIC_KEY: %w", err)
	}
	if err := viper.BindEnv("secret_key", "LANGFUSE_SECRET_KEY"); err != nil {
		return fmt.Errorf("error binding LANGFUSE_SECRET_KEY: %w", err)
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
// Currently only validates ServerUrl and base64Token as required fields.
func validateConfig(config *Config) error {
	if config.ServerUrl == "" {
		return fmt.Errorf("LANGFUSE_SERVER_URL is required")
	}

	if config.PublicKey == "" {
		return fmt.Errorf("LANGFUSE_PUBLIC_KEY is required")
	}

	if config.SecretKey == "" {
		return fmt.Errorf("LANGFUSE_SECRET_KEY is required")
	}

	if config.PublicKey != "" && config.SecretKey != "" {
		config.Base64Token = base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", config.PublicKey, config.SecretKey)))
	}

	return nil
}
