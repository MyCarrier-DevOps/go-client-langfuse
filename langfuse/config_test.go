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
	// Reset the global config variable
	config = Config{}
}

func TestLoadConfig_Success(t *testing.T) {
	defer resetViper()

	// Set environment variables
	os.Setenv("LANGFUSE_SERVER_URL", "https://test.langfuse.com")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "test-public-key")
	os.Setenv("LANGFUSE_SECRET_KEY", "test-secret-key")

	err := LoadConfig()
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
	if config.ApiToken != expectedToken {
		t.Errorf("Expected ApiToken '%s', got '%s'", expectedToken, config.ApiToken)
	}
}

func TestLoadConfig_MissingServerUrl(t *testing.T) {
	defer resetViper()

	// Explicitly unset LANGFUSE_SERVER_URL to ensure it's not set
	os.Unsetenv("LANGFUSE_SERVER_URL")
	// Set only some environment variables
	os.Setenv("LANGFUSE_PUBLIC_KEY", "test-public-key")
	os.Setenv("LANGFUSE_SECRET_KEY", "test-secret-key")

	err := LoadConfig()
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

	err := LoadConfig()
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

	err := LoadConfig()
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

	err := LoadConfig()
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
	if testConfig.ApiToken != expectedToken {
		t.Errorf("Expected ApiToken '%s', got '%s'", expectedToken, testConfig.ApiToken)
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

func TestApiTokenGeneration(t *testing.T) {
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

			if testConfig.ApiToken != tt.expected {
				t.Errorf("Expected ApiToken '%s', got '%s'", tt.expected, testConfig.ApiToken)
			}

			// Verify the token can be decoded back
			decoded, err := base64.StdEncoding.DecodeString(testConfig.ApiToken)
			if err != nil {
				t.Fatalf("Failed to decode ApiToken: %v", err)
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

			err := LoadConfig()
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
