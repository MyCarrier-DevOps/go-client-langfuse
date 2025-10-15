package langfuse

import (
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
		apiToken:        "dGVzdC1wdWJsaWMta2V5OnRlc3Qtc2VjcmV0LWtleQ==", // base64 encoded "test-public-key:test-secret-key"
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
		apiToken:        "", // Empty token
	}

	_, err := client.Do("GET", "/test")
	if err == nil {
		t.Fatal("Expected error for missing API token, got nil")
	}

	expectedError := "API token is required"
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
		ServerUrl: "https://test.langfuse.com",
		ApiToken:  "test-token",
	}

	client := NewClient()

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.baseUrl != config.ServerUrl {
		t.Errorf("Expected baseUrl %s, got %s", config.ServerUrl, client.baseUrl)
	}

	if client.apiToken != config.ApiToken {
		t.Errorf("Expected apiToken %s, got %s", config.ApiToken, client.apiToken)
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
