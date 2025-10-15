package langfuse

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

func setupProjectsTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 1
	retryClient.RetryWaitMin = 1 * time.Millisecond
	retryClient.RetryWaitMax = 10 * time.Millisecond
	retryClient.Logger = nil

	client := &Client{
		retryableClient: retryClient,
		baseUrl:         server.URL,
		apiToken:        "test-token",
	}

	client.Projects = (*ProjectsService)(&service{client: client})

	return client, server
}

func TestProjectsService_GetProject_Success(t *testing.T) {
	expectedProject := map[string]interface{}{
		"id":   "project-123",
		"name": "Test Project",
		"members": []interface{}{
			map[string]interface{}{
				"userId": "user-1",
				"role":   "admin",
			},
		},
		"createdAt": "2024-01-01T00:00:00Z",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if r.URL.Path != "/api/public/projects" {
			t.Errorf("Expected path /api/public/projects, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Authorization") == "" {
			t.Error("Expected Authorization header to be set")
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedProject)
	}

	client, server := setupProjectsTestClient(handler)
	defer server.Close()

	project, err := client.Projects.GetProject()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if project == nil {
		t.Fatal("Expected project data, got nil")
	}

	// Verify project data
	if project["id"] != "project-123" {
		t.Errorf("Expected id 'project-123', got %v", project["id"])
	}

	if project["name"] != "Test Project" {
		t.Errorf("Expected name 'Test Project', got %v", project["name"])
	}

	// Verify members array exists
	members, ok := project["members"].([]interface{})
	if !ok {
		t.Fatal("Expected members to be an array")
	}

	if len(members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(members))
	}
}

func TestProjectsService_GetProject_Unauthorized(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
	}

	client, server := setupProjectsTestClient(handler)
	defer server.Close()

	_, err := client.Projects.GetProject()
	if err == nil {
		t.Fatal("Expected error for unauthorized request, got nil")
	}

	// GetProject wraps the error with "error fetching project:"
	expectedError := "error fetching project: client error 401: unauthorized"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestProjectsService_GetProject_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("project not found"))
	}

	client, server := setupProjectsTestClient(handler)
	defer server.Close()

	_, err := client.Projects.GetProject()
	if err == nil {
		t.Fatal("Expected error for not found project, got nil")
	}

	// GetProject wraps the error with "error fetching project:"
	expectedError := "error fetching project: client error 404: project not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestProjectsService_GetProject_ServerError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}

	client, server := setupProjectsTestClient(handler)
	defer server.Close()

	_, err := client.Projects.GetProject()
	if err == nil {
		t.Fatal("Expected error for server error, got nil")
	}

	// GetProject wraps errors, and the retryable client will retry 5xx errors
	// Check that error starts with "error fetching project: error making request"
	expectedPrefix := "error fetching project: error making request"
	if err.Error()[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Expected error to start with '%s', got '%s'", expectedPrefix, err.Error())
	}
}

func TestProjectsService_GetProject_InvalidJSON(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json {{{"))
	}

	client, server := setupProjectsTestClient(handler)
	defer server.Close()

	_, err := client.Projects.GetProject()
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	// Check that error is about unmarshalling
	if err.Error()[:len("error unmarshalling")] != "error unmarshalling" {
		t.Errorf("Expected unmarshalling error, got: %v", err)
	}
}

func TestProjectsService_GetProject_EmptyResponse(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}

	client, server := setupProjectsTestClient(handler)
	defer server.Close()

	project, err := client.Projects.GetProject()
	if err != nil {
		t.Fatalf("Expected no error for empty response, got %v", err)
	}

	if project == nil {
		t.Fatal("Expected empty project map, got nil")
	}

	if len(project) != 0 {
		t.Errorf("Expected empty project map, got %d keys", len(project))
	}
}

func TestProjectsService_GetProject_ComplexStructure(t *testing.T) {
	expectedProject := map[string]interface{}{
		"id":   "project-456",
		"name": "Complex Project",
		"settings": map[string]interface{}{
			"theme": "dark",
			"notifications": map[string]interface{}{
				"email":   true,
				"slack":   false,
				"webhook": "https://example.com/webhook",
			},
		},
		"tags": []interface{}{"production", "customer-facing"},
		"metadata": map[string]interface{}{
			"department":  "engineering",
			"cost_center": 1234,
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedProject)
	}

	client, server := setupProjectsTestClient(handler)
	defer server.Close()

	project, err := client.Projects.GetProject()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify nested structures
	settings, ok := project["settings"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected settings to be a map")
	}

	if settings["theme"] != "dark" {
		t.Errorf("Expected theme 'dark', got %v", settings["theme"])
	}

	notifications, ok := settings["notifications"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected notifications to be a map")
	}

	if notifications["email"] != true {
		t.Errorf("Expected email notifications true, got %v", notifications["email"])
	}

	// Verify tags array
	tags, ok := project["tags"].([]interface{})
	if !ok {
		t.Fatal("Expected tags to be an array")
	}

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
}
