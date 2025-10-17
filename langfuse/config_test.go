package langfuse

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
)

// Helper function to reset viper and environment variables
func resetViper() {
	viper.Reset()
	os.Unsetenv("LANGFUSE_SERVER_URL")
	os.Unsetenv("LANGFUSE_PUBLIC_KEY")
	os.Unsetenv("LANGFUSE_SECRET_KEY")
}

func TestLoadConfig_Success(t *testing.T) {
	defer resetViper()

	// Set environment variables
	os.Setenv("LANGFUSE_SERVER_URL", "https://test.langfuse.com")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "test-public-key")
	os.Setenv("LANGFUSE_SECRET_KEY", "test-secret-key")

	config, err := LoadConfigFromEnvVars()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config.ServerUrl != "https://test.langfuse.com" {
		t.Errorf("Expected ServerUrl 'https://test.langfuse.com', got %s", config.ServerUrl)
	}

	if config.PublicKey != "test-public-key" {
		t.Errorf("Expected PublicKey 'test-public-key', got %s", config.PublicKey)
	}

	if config.SecretKey != "test-secret-key" {
		t.Errorf("Expected SecretKey 'test-secret-key', got %s", config.SecretKey)
	}

	// Verify API token is generated correctly
	expectedToken := base64.StdEncoding.EncodeToString([]byte("test-public-key:test-secret-key"))
	if config.Base64Token != expectedToken {
		t.Errorf("Expected Base64Token '%s', got '%s'", expectedToken, config.Base64Token)
	}
}

func TestLoadConfig_MissingServerUrl(t *testing.T) {
	defer resetViper()

	// Explicitly unset LANGFUSE_SERVER_URL to ensure it's not set
	os.Unsetenv("LANGFUSE_SERVER_URL")
	// Set only some environment variables
	os.Setenv("LANGFUSE_PUBLIC_KEY", "test-public-key")
	os.Setenv("LANGFUSE_SECRET_KEY", "test-secret-key")

	config, err := LoadConfigFromEnvVars()
	if err == nil {
		t.Fatalf("Expected error for missing LANGFUSE_SERVER_URL, got nil. Config: %+v", config)
	}

	expectedError := "error validating config: LANGFUSE_SERVER_URL is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestLoadConfig_MissingPublicKey(t *testing.T) {
	defer resetViper()

	os.Unsetenv("LANGFUSE_PUBLIC_KEY")
	os.Setenv("LANGFUSE_SERVER_URL", "https://test.langfuse.com")
	os.Setenv("LANGFUSE_SECRET_KEY", "test-secret-key")

	config, err := LoadConfigFromEnvVars()
	if err == nil {
		t.Fatalf("Expected error for missing LANGFUSE_PUBLIC_KEY, got nil. Config: %+v", config)
	}

	expectedError := "error validating config: LANGFUSE_PUBLIC_KEY is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestLoadConfig_MissingSecretKey(t *testing.T) {
	defer resetViper()

	os.Unsetenv("LANGFUSE_SECRET_KEY")
	os.Setenv("LANGFUSE_SERVER_URL", "https://test.langfuse.com")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "test-public-key")

	config, err := LoadConfigFromEnvVars()
	if err == nil {
		t.Fatalf("Expected error for missing LANGFUSE_SECRET_KEY, got nil. Config: %+v", config)
	}

	expectedError := "error validating config: LANGFUSE_SECRET_KEY is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestLoadConfig_MissingAllVariables(t *testing.T) {
	defer resetViper()

	// Explicitly unset all required environment variables
	os.Unsetenv("LANGFUSE_SERVER_URL")
	os.Unsetenv("LANGFUSE_PUBLIC_KEY")
	os.Unsetenv("LANGFUSE_SECRET_KEY")

	config, err := LoadConfigFromEnvVars()
	if err == nil {
		t.Fatalf("Expected error for missing environment variables, got nil. Config: %+v", config)
	}

	// Should fail on the first required field
	if err.Error()[:len("error validating config")] != "error validating config" {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestValidateConfig_Success(t *testing.T) {
	testConfig := &Config{
		ServerUrl: "https://test.langfuse.com",
		PublicKey: "test-public-key",
		SecretKey: "test-secret-key",
	}

	err := validateConfig(testConfig)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify API token was generated
	expectedToken := base64.StdEncoding.EncodeToString([]byte("test-public-key:test-secret-key"))
	if testConfig.Base64Token != expectedToken {
		t.Errorf("Expected Base64Token '%s', got '%s'", expectedToken, testConfig.Base64Token)
	}
}

func TestValidateConfig_EmptyServerUrl(t *testing.T) {
	testConfig := &Config{
		ServerUrl: "",
		PublicKey: "test-public-key",
		SecretKey: "test-secret-key",
	}

	err := validateConfig(testConfig)
	if err == nil {
		t.Fatal("Expected error for empty ServerUrl, got nil")
	}

	expectedError := "LANGFUSE_SERVER_URL is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestValidateConfig_EmptyPublicKey(t *testing.T) {
	testConfig := &Config{
		ServerUrl: "https://test.langfuse.com",
		PublicKey: "",
		SecretKey: "test-secret-key",
	}

	err := validateConfig(testConfig)
	if err == nil {
		t.Fatal("Expected error for empty PublicKey, got nil")
	}

	expectedError := "LANGFUSE_PUBLIC_KEY is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestValidateConfig_EmptySecretKey(t *testing.T) {
	testConfig := &Config{
		ServerUrl: "https://test.langfuse.com",
		PublicKey: "test-public-key",
		SecretKey: "",
	}

	err := validateConfig(testConfig)
	if err == nil {
		t.Fatal("Expected error for empty SecretKey, got nil")
	}

	expectedError := "LANGFUSE_SECRET_KEY is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestValidateConfig_AllFieldsEmpty(t *testing.T) {
	testConfig := &Config{
		ServerUrl: "",
		PublicKey: "",
		SecretKey: "",
	}

	err := validateConfig(testConfig)
	if err == nil {
		t.Fatal("Expected error for all empty fields, got nil")
	}

	// Should fail on the first required field (ServerUrl)
	expectedError := "LANGFUSE_SERVER_URL is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestBase64TokenGeneration(t *testing.T) {
	tests := []struct {
		name      string
		publicKey string
		secretKey string
		expected  string
	}{
		{
			name:      "Simple keys",
			publicKey: "public",
			secretKey: "secret",
			expected:  base64.StdEncoding.EncodeToString([]byte("public:secret")),
		},
		{
			name:      "Keys with special characters",
			publicKey: "pk-test-123",
			secretKey: "sk-test-456",
			expected:  base64.StdEncoding.EncodeToString([]byte("pk-test-123:sk-test-456")),
		},
		{
			name:      "Long keys",
			publicKey: "pk-very-long-public-key-with-many-characters-12345",
			secretKey: "sk-very-long-secret-key-with-many-characters-67890",
			expected:  base64.StdEncoding.EncodeToString([]byte("pk-very-long-public-key-with-many-characters-12345:sk-very-long-secret-key-with-many-characters-67890")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testConfig := &Config{
				ServerUrl: "https://test.langfuse.com",
				PublicKey: tt.publicKey,
				SecretKey: tt.secretKey,
			}

			err := validateConfig(testConfig)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if testConfig.Base64Token != tt.expected {
				t.Errorf("Expected Base64Token '%s', got '%s'", tt.expected, testConfig.Base64Token)
			}

			// Verify the token can be decoded back
			decoded, err := base64.StdEncoding.DecodeString(testConfig.Base64Token)
			if err != nil {
				t.Fatalf("Failed to decode Base64Token: %v", err)
			}

			expectedDecoded := fmt.Sprintf("%s:%s", tt.publicKey, tt.secretKey)
			if string(decoded) != expectedDecoded {
				t.Errorf("Expected decoded token '%s', got '%s'", expectedDecoded, string(decoded))
			}
		})
	}
}

func TestLoadConfig_WithDifferentServerUrls(t *testing.T) {
	tests := []struct {
		name      string
		serverUrl string
	}{
		{
			name:      "Cloud URL",
			serverUrl: "https://cloud.langfuse.com",
		},
		{
			name:      "US Cloud URL",
			serverUrl: "https://us.cloud.langfuse.com",
		},
		{
			name:      "EU Cloud URL",
			serverUrl: "https://eu.cloud.langfuse.com",
		},
		{
			name:      "Self-hosted URL",
			serverUrl: "https://langfuse.example.com",
		},
		{
			name:      "Localhost URL",
			serverUrl: "http://localhost:3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetViper()

			os.Setenv("LANGFUSE_SERVER_URL", tt.serverUrl)
			os.Setenv("LANGFUSE_PUBLIC_KEY", "test-public-key")
			os.Setenv("LANGFUSE_SECRET_KEY", "test-secret-key")

			config, err := LoadConfigFromEnvVars()
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if config.ServerUrl != tt.serverUrl {
				t.Errorf("Expected ServerUrl '%s', got '%s'", tt.serverUrl, config.ServerUrl)
			}
		})
	}
}

func TestConfig_StructTags(t *testing.T) {
	defer resetViper()

	// Set environment variables
	os.Setenv("LANGFUSE_SERVER_URL", "https://test.langfuse.com")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "test-public-key")
	os.Setenv("LANGFUSE_SECRET_KEY", "test-secret-key")

	viper.BindEnv("server_url", "LANGFUSE_SERVER_URL")
	viper.BindEnv("public_key", "LANGFUSE_PUBLIC_KEY")
	viper.BindEnv("secret_key", "LANGFUSE_SECRET_KEY")
	viper.AutomaticEnv()

	var testConfig Config
	err := viper.Unmarshal(&testConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify mapstructure tags work correctly
	if testConfig.ServerUrl != "https://test.langfuse.com" {
		t.Errorf("Expected ServerUrl to be set via mapstructure tag")
	}

	if testConfig.PublicKey != "test-public-key" {
		t.Errorf("Expected PublicKey to be set via mapstructure tag")
	}

	if testConfig.SecretKey != "test-secret-key" {
		t.Errorf("Expected SecretKey to be set via mapstructure tag")
	}
}

func TestNewConfig_Success(t *testing.T) {
	cfg, err := NewConfig(
		"https://cloud.langfuse.com",
		"pk-lf-test-123",
		"sk-lf-test-456",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.ServerUrl != "https://cloud.langfuse.com" {
		t.Errorf("Expected ServerUrl 'https://cloud.langfuse.com', got %s", cfg.ServerUrl)
	}

	if cfg.PublicKey != "pk-lf-test-123" {
		t.Errorf("Expected PublicKey 'pk-lf-test-123', got %s", cfg.PublicKey)
	}

	if cfg.SecretKey != "sk-lf-test-456" {
		t.Errorf("Expected SecretKey 'sk-lf-test-456', got %s", cfg.SecretKey)
	}

	// Verify Base64Token was generated
	expectedToken := base64.StdEncoding.EncodeToString([]byte("pk-lf-test-123:sk-lf-test-456"))
	if cfg.Base64Token != expectedToken {
		t.Errorf("Expected Base64Token '%s', got '%s'", expectedToken, cfg.Base64Token)
	}
}

func TestNewConfig_MissingServerUrl(t *testing.T) {
	_, err := NewConfig("", "pk-lf-test-123", "sk-lf-test-456")

	if err == nil {
		t.Fatal("Expected error for empty ServerUrl, got nil")
	}

	expectedError := "LANGFUSE_SERVER_URL is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestNewConfig_MissingPublicKey(t *testing.T) {
	_, err := NewConfig("https://cloud.langfuse.com", "", "sk-lf-test-456")

	if err == nil {
		t.Fatal("Expected error for empty PublicKey, got nil")
	}

	expectedError := "LANGFUSE_PUBLIC_KEY is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestNewConfig_MissingSecretKey(t *testing.T) {
	_, err := NewConfig("https://cloud.langfuse.com", "pk-lf-test-123", "")

	if err == nil {
		t.Fatal("Expected error for empty SecretKey, got nil")
	}

	expectedError := "LANGFUSE_SECRET_KEY is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestNewConfig_AllFieldsEmpty(t *testing.T) {
	_, err := NewConfig("", "", "")

	if err == nil {
		t.Fatal("Expected error for all empty fields, got nil")
	}

	// Should fail on the first required field (ServerUrl)
	expectedError := "LANGFUSE_SERVER_URL is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestNewConfig_WithDifferentURLFormats(t *testing.T) {
	tests := []struct {
		name      string
		serverUrl string
		publicKey string
		secretKey string
	}{
		{
			name:      "Cloud URL",
			serverUrl: "https://cloud.langfuse.com",
			publicKey: "pk-lf-test",
			secretKey: "sk-lf-test",
		},
		{
			name:      "US Cloud URL",
			serverUrl: "https://us.cloud.langfuse.com",
			publicKey: "pk-lf-us-test",
			secretKey: "sk-lf-us-test",
		},
		{
			name:      "EU Cloud URL",
			serverUrl: "https://eu.cloud.langfuse.com",
			publicKey: "pk-lf-eu-test",
			secretKey: "sk-lf-eu-test",
		},
		{
			name:      "Self-hosted URL",
			serverUrl: "https://langfuse.example.com",
			publicKey: "pk-custom-test",
			secretKey: "sk-custom-test",
		},
		{
			name:      "Localhost URL",
			serverUrl: "http://localhost:3000",
			publicKey: "pk-local-test",
			secretKey: "sk-local-test",
		},
		{
			name:      "URL with port",
			serverUrl: "https://langfuse.example.com:8080",
			publicKey: "pk-port-test",
			secretKey: "sk-port-test",
		},
		{
			name:      "URL with path",
			serverUrl: "https://example.com/api/langfuse",
			publicKey: "pk-path-test",
			secretKey: "sk-path-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewConfig(tt.serverUrl, tt.publicKey, tt.secretKey)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if cfg.ServerUrl != tt.serverUrl {
				t.Errorf("Expected ServerUrl '%s', got '%s'", tt.serverUrl, cfg.ServerUrl)
			}

			if cfg.PublicKey != tt.publicKey {
				t.Errorf("Expected PublicKey '%s', got '%s'", tt.publicKey, cfg.PublicKey)
			}

			if cfg.SecretKey != tt.secretKey {
				t.Errorf("Expected SecretKey '%s', got '%s'", tt.secretKey, cfg.SecretKey)
			}

			// Verify Base64Token was generated correctly
			expectedToken := base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", tt.publicKey, tt.secretKey)))
			if cfg.Base64Token != expectedToken {
				t.Errorf("Expected Base64Token '%s', got '%s'", expectedToken, cfg.Base64Token)
			}
		})
	}
}

func TestNewConfig_WithSpecialCharactersInKeys(t *testing.T) {
	tests := []struct {
		name      string
		publicKey string
		secretKey string
	}{
		{
			name:      "Keys with dashes",
			publicKey: "pk-lf-test-123-456",
			secretKey: "sk-lf-test-789-012",
		},
		{
			name:      "Keys with underscores",
			publicKey: "pk_lf_test_123",
			secretKey: "sk_lf_test_456",
		},
		{
			name:      "Long keys",
			publicKey: "pk-very-long-public-key-with-many-segments-123456789",
			secretKey: "sk-very-long-secret-key-with-many-segments-987654321",
		},
		{
			name:      "Keys with alphanumeric",
			publicKey: "pkABC123xyz",
			secretKey: "skXYZ789abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewConfig("https://cloud.langfuse.com", tt.publicKey, tt.secretKey)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if cfg.PublicKey != tt.publicKey {
				t.Errorf("Expected PublicKey '%s', got '%s'", tt.publicKey, cfg.PublicKey)
			}

			if cfg.SecretKey != tt.secretKey {
				t.Errorf("Expected SecretKey '%s', got '%s'", tt.secretKey, cfg.SecretKey)
			}

			// Verify Base64Token was generated correctly
			expectedToken := base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", tt.publicKey, tt.secretKey)))
			if cfg.Base64Token != expectedToken {
				t.Errorf("Expected Base64Token '%s', got '%s'", expectedToken, cfg.Base64Token)
			}

			// Verify the token can be decoded
			decoded, err := base64.StdEncoding.DecodeString(cfg.Base64Token)
			if err != nil {
				t.Fatalf("Failed to decode Base64Token: %v", err)
			}

			expectedDecoded := fmt.Sprintf("%s:%s", tt.publicKey, tt.secretKey)
			if string(decoded) != expectedDecoded {
				t.Errorf("Expected decoded token '%s', got '%s'", expectedDecoded, string(decoded))
			}
		})
	}
}

func TestConfigCreation_BothMethods(t *testing.T) {
	// Test that both config creation methods produce equivalent results
	defer resetViper()

	testServerUrl := "https://test.langfuse.com"
	testPublicKey := "pk-test-123"
	testSecretKey := "sk-test-456"

	// Method 1: Using environment variables
	os.Setenv("LANGFUSE_SERVER_URL", testServerUrl)
	os.Setenv("LANGFUSE_PUBLIC_KEY", testPublicKey)
	os.Setenv("LANGFUSE_SECRET_KEY", testSecretKey)

	config, err := LoadConfigFromEnvVars()
	if err != nil {
		t.Fatalf("LoadConfigFromEnvVars failed: %v", err)
	}

	envConfig := config

	// Method 2: Using NewConfig
	directConfig, err := NewConfig(testServerUrl, testPublicKey, testSecretKey)
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	// Compare both configs
	if envConfig.ServerUrl != directConfig.ServerUrl {
		t.Errorf("ServerUrl mismatch: env=%s, direct=%s", envConfig.ServerUrl, directConfig.ServerUrl)
	}

	if envConfig.PublicKey != directConfig.PublicKey {
		t.Errorf("PublicKey mismatch: env=%s, direct=%s", envConfig.PublicKey, directConfig.PublicKey)
	}

	if envConfig.SecretKey != directConfig.SecretKey {
		t.Errorf("SecretKey mismatch: env=%s, direct=%s", envConfig.SecretKey, directConfig.SecretKey)
	}

	if envConfig.Base64Token != directConfig.Base64Token {
		t.Errorf("Base64Token mismatch: env=%s, direct=%s", envConfig.Base64Token, directConfig.Base64Token)
	}
}

func TestNewConfig_ReusabilityAndIsolation(t *testing.T) {
	// Test that multiple configs can be created independently
	cfg1, err := NewConfig(
		"https://server1.com",
		"pk-1",
		"sk-1",
	)
	if err != nil {
		t.Fatalf("Failed to create first config: %v", err)
	}

	cfg2, err := NewConfig(
		"https://server2.com",
		"pk-2",
		"sk-2",
	)
	if err != nil {
		t.Fatalf("Failed to create second config: %v", err)
	}

	// Verify configs are independent
	if cfg1.ServerUrl == cfg2.ServerUrl {
		t.Error("Configs should have different ServerUrls")
	}

	if cfg1.PublicKey == cfg2.PublicKey {
		t.Error("Configs should have different PublicKeys")
	}

	if cfg1.SecretKey == cfg2.SecretKey {
		t.Error("Configs should have different SecretKeys")
	}

	if cfg1.Base64Token == cfg2.Base64Token {
		t.Error("Configs should have different Base64Tokens")
	}

	// Verify first config is still correct
	if cfg1.ServerUrl != "https://server1.com" {
		t.Errorf("First config ServerUrl changed unexpectedly: %s", cfg1.ServerUrl)
	}

	// Verify second config is correct
	if cfg2.ServerUrl != "https://server2.com" {
		t.Errorf("Second config ServerUrl incorrect: %s", cfg2.ServerUrl)
	}
}

func TestLoadConfigFromEnvVars_MultipleCalls(t *testing.T) {
	defer resetViper()

	// First call - set initial values
	os.Setenv("LANGFUSE_SERVER_URL", "https://first.langfuse.com")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "pk-first")
	os.Setenv("LANGFUSE_SECRET_KEY", "sk-first")

	config, err := LoadConfigFromEnvVars()
	if err != nil {
		t.Fatalf("First LoadConfigFromEnvVars failed: %v", err)
	}

	firstToken := config.Base64Token

	// Second call - update values
	os.Setenv("LANGFUSE_SERVER_URL", "https://second.langfuse.com")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "pk-second")
	os.Setenv("LANGFUSE_SECRET_KEY", "sk-second")

	// Reset viper to pick up new values
	viper.Reset()
	config, err = LoadConfigFromEnvVars()
	if err != nil {
		t.Fatalf("Second LoadConfigFromEnvVars failed: %v", err)
	}

	// Verify config was updated
	if config.ServerUrl != "https://second.langfuse.com" {
		t.Errorf("Expected ServerUrl to be updated, got %s", config.ServerUrl)
	}

	if config.Base64Token == firstToken {
		t.Error("Expected Base64Token to be updated")
	}
}

func TestLoadConfigFromEnvVars_WithWhitespace(t *testing.T) {
	defer resetViper()

	// Set environment variables with leading/trailing whitespace
	os.Setenv("LANGFUSE_SERVER_URL", "  https://test.langfuse.com  ")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "  pk-test  ")
	os.Setenv("LANGFUSE_SECRET_KEY", "  sk-test  ")

	config, err := LoadConfigFromEnvVars()
	if err != nil {
		t.Fatalf("LoadConfigFromEnvVars failed: %v", err)
	}

	// Note: viper doesn't automatically trim whitespace, so we're testing actual behavior
	// The config will contain the whitespace as-is
	if config.ServerUrl == "" {
		t.Error("Expected ServerUrl to be set")
	}

	if config.PublicKey == "" {
		t.Error("Expected PublicKey to be set")
	}

	if config.SecretKey == "" {
		t.Error("Expected SecretKey to be set")
	}
}

func TestLoadConfigFromEnvVars_EmptyStringVsUnset(t *testing.T) {
	defer resetViper()

	// Test with explicitly empty string (not unset)
	os.Setenv("LANGFUSE_SERVER_URL", "")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "pk-test")
	os.Setenv("LANGFUSE_SECRET_KEY", "sk-test")

	_, err := LoadConfigFromEnvVars()
	if err == nil {
		t.Fatal("Expected error for empty ServerUrl, got nil")
	}

	expectedError := "error validating config: LANGFUSE_SERVER_URL is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestLoadConfigFromEnvVars_PartialConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		serverUrl     string
		publicKey     string
		secretKey     string
		expectedError string
	}{
		{
			name:          "Only ServerUrl set",
			serverUrl:     "https://test.com",
			publicKey:     "",
			secretKey:     "",
			expectedError: "error validating config: LANGFUSE_PUBLIC_KEY is required",
		},
		{
			name:          "Only PublicKey set",
			serverUrl:     "",
			publicKey:     "pk-test",
			secretKey:     "",
			expectedError: "error validating config: LANGFUSE_SERVER_URL is required",
		},
		{
			name:          "Only SecretKey set",
			serverUrl:     "",
			publicKey:     "",
			secretKey:     "sk-test",
			expectedError: "error validating config: LANGFUSE_SERVER_URL is required",
		},
		{
			name:          "ServerUrl and PublicKey only",
			serverUrl:     "https://test.com",
			publicKey:     "pk-test",
			secretKey:     "",
			expectedError: "error validating config: LANGFUSE_SECRET_KEY is required",
		},
		{
			name:          "ServerUrl and SecretKey only",
			serverUrl:     "https://test.com",
			publicKey:     "",
			secretKey:     "sk-test",
			expectedError: "error validating config: LANGFUSE_PUBLIC_KEY is required",
		},
		{
			name:          "PublicKey and SecretKey only",
			serverUrl:     "",
			publicKey:     "pk-test",
			secretKey:     "sk-test",
			expectedError: "error validating config: LANGFUSE_SERVER_URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetViper()

			if tt.serverUrl != "" {
				os.Setenv("LANGFUSE_SERVER_URL", tt.serverUrl)
			} else {
				os.Unsetenv("LANGFUSE_SERVER_URL")
			}

			if tt.publicKey != "" {
				os.Setenv("LANGFUSE_PUBLIC_KEY", tt.publicKey)
			} else {
				os.Unsetenv("LANGFUSE_PUBLIC_KEY")
			}

			if tt.secretKey != "" {
				os.Setenv("LANGFUSE_SECRET_KEY", tt.secretKey)
			} else {
				os.Unsetenv("LANGFUSE_SECRET_KEY")
			}

			_, err := LoadConfigFromEnvVars()
			if err == nil {
				t.Fatalf("Expected error '%s', got nil", tt.expectedError)
			}

			if err.Error() != tt.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

func TestLoadConfigFromEnvVars_CasePreservation(t *testing.T) {
	defer resetViper()

	// Test that case is preserved in keys
	os.Setenv("LANGFUSE_SERVER_URL", "https://test.langfuse.com")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "PK-MixedCase-123")
	os.Setenv("LANGFUSE_SECRET_KEY", "SK-MixedCase-456")

	config, err := LoadConfigFromEnvVars()
	if err != nil {
		t.Fatalf("LoadConfigFromEnvVars failed: %v", err)
	}

	if config.PublicKey != "PK-MixedCase-123" {
		t.Errorf("Expected PublicKey 'PK-MixedCase-123', got '%s'", config.PublicKey)
	}

	if config.SecretKey != "SK-MixedCase-456" {
		t.Errorf("Expected SecretKey 'SK-MixedCase-456', got '%s'", config.SecretKey)
	}

	// Verify Base64Token contains the mixed case keys
	expectedToken := base64.StdEncoding.EncodeToString([]byte("PK-MixedCase-123:SK-MixedCase-456"))
	if config.Base64Token != expectedToken {
		t.Errorf("Expected Base64Token '%s', got '%s'", expectedToken, config.Base64Token)
	}
}

func TestLoadConfigFromEnvVars_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		setupEnv      func()
		expectedError string
	}{
		{
			name: "Missing ServerUrl validation",
			setupEnv: func() {
				os.Unsetenv("LANGFUSE_SERVER_URL")
				os.Setenv("LANGFUSE_PUBLIC_KEY", "pk-test")
				os.Setenv("LANGFUSE_SECRET_KEY", "sk-test")
			},
			expectedError: "error validating config: LANGFUSE_SERVER_URL is required",
		},
		{
			name: "Missing PublicKey validation",
			setupEnv: func() {
				os.Setenv("LANGFUSE_SERVER_URL", "https://test.com")
				os.Unsetenv("LANGFUSE_PUBLIC_KEY")
				os.Setenv("LANGFUSE_SECRET_KEY", "sk-test")
			},
			expectedError: "error validating config: LANGFUSE_PUBLIC_KEY is required",
		},
		{
			name: "Missing SecretKey validation",
			setupEnv: func() {
				os.Setenv("LANGFUSE_SERVER_URL", "https://test.com")
				os.Setenv("LANGFUSE_PUBLIC_KEY", "pk-test")
				os.Unsetenv("LANGFUSE_SECRET_KEY")
			},
			expectedError: "error validating config: LANGFUSE_SECRET_KEY is required",
		},
		{
			name: "All fields missing",
			setupEnv: func() {
				os.Unsetenv("LANGFUSE_SERVER_URL")
				os.Unsetenv("LANGFUSE_PUBLIC_KEY")
				os.Unsetenv("LANGFUSE_SECRET_KEY")
			},
			expectedError: "error validating config: LANGFUSE_SERVER_URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetViper()

			tt.setupEnv()

			_, err := LoadConfigFromEnvVars()
			if err == nil {
				t.Fatalf("Expected error '%s', got nil", tt.expectedError)
			}

			if err.Error() != tt.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

func TestLoadConfigFromEnvVars_ErrorMessageFormat(t *testing.T) {
	defer resetViper()

	// Test that error messages are properly formatted with context
	os.Unsetenv("LANGFUSE_SERVER_URL")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "pk-test")
	os.Setenv("LANGFUSE_SECRET_KEY", "sk-test")

	_, err := LoadConfigFromEnvVars()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error message contains expected format
	expectedPrefix := "error validating config:"
	if len(err.Error()) < len(expectedPrefix) || err.Error()[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Expected error to start with '%s', got '%s'", expectedPrefix, err.Error())
	}
}

func TestLoadConfigFromEnvVars_SuccessfulValidation(t *testing.T) {
	defer resetViper()

	// Test complete successful flow
	os.Setenv("LANGFUSE_SERVER_URL", "https://cloud.langfuse.com")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "pk-lf-valid-key-123")
	os.Setenv("LANGFUSE_SECRET_KEY", "sk-lf-valid-key-456")

	config, err := LoadConfigFromEnvVars()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify all fields are correctly set
	if config.ServerUrl != "https://cloud.langfuse.com" {
		t.Errorf("ServerUrl not set correctly")
	}

	if config.PublicKey != "pk-lf-valid-key-123" {
		t.Errorf("PublicKey not set correctly")
	}

	if config.SecretKey != "sk-lf-valid-key-456" {
		t.Errorf("SecretKey not set correctly")
	}

	// Verify Base64Token was generated
	if config.Base64Token == "" {
		t.Error("Base64Token should be generated")
	}

	// Verify Base64Token is valid
	expectedToken := base64.StdEncoding.EncodeToString(
		[]byte("pk-lf-valid-key-123:sk-lf-valid-key-456"))
	if config.Base64Token != expectedToken {
		t.Errorf("Base64Token not generated correctly. Expected '%s', got '%s'",
			expectedToken, config.Base64Token)
	}
}

func TestLoadConfigFromEnvVars_SequentialCallsWithDifferentValues(t *testing.T) {
	defer resetViper()

	scenarios := []struct {
		serverUrl string
		publicKey string
		secretKey string
	}{
		{
			serverUrl: "https://server1.com",
			publicKey: "pk-1",
			secretKey: "sk-1",
		},
		{
			serverUrl: "https://server2.com",
			publicKey: "pk-2",
			secretKey: "sk-2",
		},
		{
			serverUrl: "https://server3.com",
			publicKey: "pk-3",
			secretKey: "sk-3",
		},
	}

	for i, scenario := range scenarios {
		t.Run(fmt.Sprintf("Iteration_%d", i+1), func(t *testing.T) {
			// Reset viper before each scenario
			viper.Reset()

			os.Setenv("LANGFUSE_SERVER_URL", scenario.serverUrl)
			os.Setenv("LANGFUSE_PUBLIC_KEY", scenario.publicKey)
			os.Setenv("LANGFUSE_SECRET_KEY", scenario.secretKey)

			config, err := LoadConfigFromEnvVars()
			if err != nil {
				t.Fatalf("LoadConfigFromEnvVars failed: %v", err)
			}

			if config.ServerUrl != scenario.serverUrl {
				t.Errorf("Expected ServerUrl '%s', got '%s'", scenario.serverUrl, config.ServerUrl)
			}

			if config.PublicKey != scenario.publicKey {
				t.Errorf("Expected PublicKey '%s', got '%s'", scenario.publicKey, config.PublicKey)
			}

			if config.SecretKey != scenario.secretKey {
				t.Errorf("Expected SecretKey '%s', got '%s'", scenario.secretKey, config.SecretKey)
			}
		})
	}
}
