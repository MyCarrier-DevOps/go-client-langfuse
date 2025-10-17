package langfuse

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// setupTestClient creates a test client with a mock server
func setupTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(handler))

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 1
	retryClient.RetryWaitMin = 1 * time.Millisecond
	retryClient.RetryWaitMax = 10 * time.Millisecond
	retryClient.Logger = nil

	client := &Client{
		retryableClient: retryClient,
		baseUrl:         server.URL,
		base64Token:     "dGVzdC1wdWJsaWMta2V5OnRlc3Qtc2VjcmV0LWtleQ==", // base64 encoded "test-public-key:test-secret-key"
	}

	client.Projects = (*ProjectsService)(&service{client: client})
	client.Prompts = (*PromptsService)(&service{client: client})

	return client, server
}

func TestClient_Do_Success(t *testing.T) {
	expectedResponse := map[string]string{"status": "ok"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "Basic dGVzdC1wdWJsaWMta2V5OnRlc3Qtc2VjcmV0LWtleQ==" {
			t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("User-Agent") != defaultUserAgent {
			t.Errorf("Expected User-Agent header %s, got %s", defaultUserAgent, r.Header.Get("User-Agent"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	body, err := client.Do("GET", "/test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var response map[string]string
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %s", response["status"])
	}
}

func TestClient_Do_MissingAPIToken(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil

	client := &Client{
		retryableClient: retryClient,
		baseUrl:         server.URL,
		base64Token:     "", // Empty token
	}

	_, err := client.Do("GET", "/test")
	if err == nil {
		t.Fatal("Expected error for missing API token, got nil")
	}

	expectedError := "Base64 token is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestClient_Do_ClientError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	_, err := client.Do("GET", "/test")
	if err == nil {
		t.Fatal("Expected error for 400 status, got nil")
	}

	expectedError := "client error 400: bad request"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestClient_Do_ServerError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	_, err := client.Do("GET", "/test")
	if err == nil {
		t.Fatal("Expected error for 500 status, got nil")
	}

	// The retryable client will exhaust retries for 5xx errors
	// Check that the error message contains "error making request" and "giving up"
	if err.Error()[:len("error making request")] != "error making request" {
		t.Errorf("Expected error to start with 'error making request', got '%s'", err.Error())
	}
}

func TestClient_DoWithBody_Success(t *testing.T) {
	expectedPayload := map[string]string{
		"name": "test",
		"type": "chat",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify payload
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if payload["name"] != expectedPayload["name"] {
			t.Errorf("Expected name %s, got %s", expectedPayload["name"], payload["name"])
		}

		if payload["type"] != expectedPayload["type"] {
			t.Errorf("Expected type %s, got %s", expectedPayload["type"], payload["type"])
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "123", "status": "created"})
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	body, err := client.DoWithBody("POST", "/test", expectedPayload)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var response map[string]string
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["id"] != "123" {
		t.Errorf("Expected id '123', got %s", response["id"])
	}

	if response["status"] != "created" {
		t.Errorf("Expected status 'created', got %s", response["status"])
	}
}

func TestClient_DoWithBody_NilPayload(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify that body is empty
		if r.ContentLength > 0 {
			t.Error("Expected empty body, got content")
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "ok"}`)
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	_, err := client.DoWithBody("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Expected no error with nil payload, got %v", err)
	}
}

func TestClient_DoWithBody_InvalidJSON(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	// Try to marshal a channel (which can't be marshaled to JSON)
	invalidPayload := make(chan int)

	_, err := client.DoWithBody("POST", "/test", invalidPayload)
	if err == nil {
		t.Fatal("Expected error for invalid JSON payload, got nil")
	}

	if err.Error()[:len("error marshalling request body")] != "error marshalling request body" {
		t.Errorf("Expected JSON marshalling error, got: %v", err)
	}
}

func TestNewClient(t *testing.T) {
	// Set up config for testing
	config = Config{
		ServerUrl:   "https://test.langfuse.com",
		Base64Token: "test-token",
	}

	client := NewClient()

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.baseUrl != config.ServerUrl {
		t.Errorf("Expected baseUrl %s, got %s", config.ServerUrl, client.baseUrl)
	}

	if client.base64Token != config.Base64Token {
		t.Errorf("Expected base64Token %s, got %s", config.Base64Token, client.base64Token)
	}

	if client.Projects == nil {
		t.Error("Expected Projects service to be initialized")
	}

	if client.Prompts == nil {
		t.Error("Expected Prompts service to be initialized")
	}

	if client.retryableClient == nil {
		t.Error("Expected retryableClient to be initialized")
	}

	if client.retryableClient.RetryMax != 3 {
		t.Errorf("Expected RetryMax 3, got %d", client.retryableClient.RetryMax)
	}
}

func TestNewClientWithConfig(t *testing.T) {
	cfg, err := NewConfig(
		"https://cloud.langfuse.com",
		"pk-lf-test-123",
		"sk-lf-test-456",
	)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	client := NewClientWithConfig(cfg)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.baseUrl != cfg.ServerUrl {
		t.Errorf("Expected baseUrl %s, got %s", cfg.ServerUrl, client.baseUrl)
	}

	if client.base64Token != cfg.Base64Token {
		t.Errorf("Expected base64Token %s, got %s", cfg.Base64Token, client.base64Token)
	}

	if client.Projects == nil {
		t.Error("Expected Projects service to be initialized")
	}

	if client.Prompts == nil {
		t.Error("Expected Prompts service to be initialized")
	}

	if client.retryableClient == nil {
		t.Error("Expected retryableClient to be initialized")
	}

	// Verify retry configuration
	if client.retryableClient.RetryMax != 3 {
		t.Errorf("Expected RetryMax 3, got %d", client.retryableClient.RetryMax)
	}

	if client.retryableClient.RetryWaitMin != 1*time.Second {
		t.Errorf("Expected RetryWaitMin 1s, got %v", client.retryableClient.RetryWaitMin)
	}

	if client.retryableClient.RetryWaitMax != 4*time.Second {
		t.Errorf("Expected RetryWaitMax 4s, got %v", client.retryableClient.RetryWaitMax)
	}
}

func TestNewClientWithConfig_CustomServerUrl(t *testing.T) {
	customUrls := []string{
		"https://cloud.langfuse.com",
		"https://us.cloud.langfuse.com",
		"https://eu.cloud.langfuse.com",
		"https://langfuse.example.com",
		"http://localhost:3000",
	}

	for _, url := range customUrls {
		t.Run(url, func(t *testing.T) {
			cfg, err := NewConfig(url, "pk-test", "sk-test")
			if err != nil {
				t.Fatalf("Failed to create config: %v", err)
			}

			client := NewClientWithConfig(cfg)

			if client.baseUrl != url {
				t.Errorf("Expected baseUrl %s, got %s", url, client.baseUrl)
			}
		})
	}
}

func TestNewClientWithConfig_WithRealRequest(t *testing.T) {
	// Create a test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify the Authorization header is set correctly
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Expected Authorization header to be set")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// Create config with test server URL
	cfg, err := NewConfig(server.URL, "pk-test", "sk-test")
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Create client with config
	client := NewClientWithConfig(cfg)

	// Make a request
	_, err = client.Do("GET", "/test")
	if err != nil {
		t.Fatalf("Expected request to succeed, got error: %v", err)
	}
}

func TestDoWithBody_DifferentHTTPMethods(t *testing.T) {
	methods := []string{"POST", "PUT", "PATCH", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					t.Errorf("Expected method %s, got %s", method, r.Method)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true}`))
			}

			client, server := setupTestClient(handler)
			defer server.Close()

			payload := map[string]string{"key": "value"}
			_, err := client.DoWithBody(method, "/test", payload)
			if err != nil {
				t.Fatalf("Expected no error for %s, got %v", method, err)
			}
		})
	}
}

func TestDoWithBody_EmptyMethod(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// When method is empty, it should default to GET
		if r.Method != "GET" {
			t.Errorf("Expected method GET (default), got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	_, err := client.DoWithBody("", "/test", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDoWithBody_LargePayload(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify large data field exists
		if _, ok := payload["large_data"]; !ok {
			t.Error("Expected large_data field in payload")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	// Create a large payload
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	payload := map[string]interface{}{
		"large_data": base64.StdEncoding.EncodeToString(largeData),
		"metadata":   "test",
	}

	_, err := client.DoWithBody("POST", "/test", payload)
	if err != nil {
		t.Fatalf("Expected no error with large payload, got %v", err)
	}
}

func TestDoWithBody_NestedJSON(t *testing.T) {
	expectedPayload := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "test user",
			"profile": map[string]interface{}{
				"age":     30,
				"country": "US",
			},
		},
		"tags":  []string{"tag1", "tag2", "tag3"},
		"count": 42,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify nested structure
		user, ok := payload["user"].(map[string]interface{})
		if !ok {
			t.Error("Expected user to be a nested object")
		} else {
			if user["name"] != "test user" {
				t.Errorf("Expected name 'test user', got %v", user["name"])
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	_, err := client.DoWithBody("POST", "/test", expectedPayload)
	if err != nil {
		t.Fatalf("Expected no error with nested JSON, got %v", err)
	}
}

func TestDoWithBody_ArrayPayload(t *testing.T) {
	expectedPayload := []map[string]string{
		{"id": "1", "name": "item1"},
		{"id": "2", "name": "item2"},
		{"id": "3", "name": "item3"},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var payload []map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(payload) != 3 {
			t.Errorf("Expected 3 items, got %d", len(payload))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	_, err := client.DoWithBody("POST", "/test", expectedPayload)
	if err != nil {
		t.Fatalf("Expected no error with array payload, got %v", err)
	}
}

func TestDoWithBody_SpecialCharactersInPayload(t *testing.T) {
	payload := map[string]string{
		"special": "Hello \"World\" with 'quotes' and \n newlines \t tabs",
		"unicode": "こんにちは世界 🌍 émojis",
		"symbols": "!@#$%^&*()_+-=[]{}|;:',.<>?/",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var received map[string]string
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if received["special"] != payload["special"] {
			t.Errorf("Special characters not preserved")
		}

		if received["unicode"] != payload["unicode"] {
			t.Errorf("Unicode characters not preserved")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	_, err := client.DoWithBody("POST", "/test", payload)
	if err != nil {
		t.Fatalf("Expected no error with special characters, got %v", err)
	}
}

func TestDoWithBody_ResponseBodyReading(t *testing.T) {
	expectedResponse := map[string]interface{}{
		"id":      "123",
		"status":  "created",
		"message": "Success",
		"data": map[string]string{
			"key": "value",
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(expectedResponse)
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	body, err := client.DoWithBody("POST", "/test", map[string]string{"test": "data"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["id"] != "123" {
		t.Errorf("Expected id '123', got %v", response["id"])
	}

	if response["status"] != "created" {
		t.Errorf("Expected status 'created', got %v", response["status"])
	}
}

func TestDoWithBody_ContentTypeHeader(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	payload := map[string]string{"test": "data"}
	_, err := client.DoWithBody("POST", "/test", payload)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDoWithBody_MultipleSequentialRequests(t *testing.T) {
	requestCount := 0

	handler := func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"request": %d}`, requestCount)))
	}

	client, server := setupTestClient(handler)
	defer server.Close()

	// Make multiple sequential requests
	for i := 1; i <= 5; i++ {
		payload := map[string]int{"request_number": i}
		_, err := client.DoWithBody("POST", "/test", payload)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}

	if requestCount != 5 {
		t.Errorf("Expected 5 requests, got %d", requestCount)
	}
}
