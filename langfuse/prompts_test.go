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

func TestPromptMarshaling(t *testing.T) {
	// Test chat prompt marshaling
	chatPrompt := &Prompt{
		Type: "chat",
		Name: "test-chat-prompt",
		Prompt: []ChatMessage{
			{
				Type:    "chatmessage",
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Type:    "chatmessage",
				Role:    "user",
				Content: "Hello!",
			},
		},
		Labels:        []string{"production", "v1"},
		Tags:          []string{"chat", "test"},
		CommitMessage: "Initial version",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(chatPrompt)
	if err != nil {
		t.Fatalf("Failed to marshal chat prompt: %v", err)
	}

	// Unmarshal back
	var unmarshaled Prompt
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal chat prompt: %v", err)
	}

	// Verify basic fields
	if unmarshaled.Type != chatPrompt.Type {
		t.Errorf("Expected type %s, got %s", chatPrompt.Type, unmarshaled.Type)
	}

	if unmarshaled.Name != chatPrompt.Name {
		t.Errorf("Expected name %s, got %s", chatPrompt.Name, unmarshaled.Name)
	}

	if unmarshaled.CommitMessage != chatPrompt.CommitMessage {
		t.Errorf("Expected commit message %s, got %s", chatPrompt.CommitMessage, unmarshaled.CommitMessage)
	}
}

func TestPromptWithNullableFields(t *testing.T) {
	// Test prompt with null config and commitMessage
	promptJSON := `{
		"type": "chat",
		"name": "test-prompt",
		"config": null,
		"commitMessage": null,
		"labels": ["production"],
		"tags": ["test"]
	}`

	var prompt Prompt
	err := json.Unmarshal([]byte(promptJSON), &prompt)
	if err != nil {
		t.Fatalf("Failed to unmarshal prompt with null fields: %v", err)
	}

	if prompt.Type != "chat" {
		t.Errorf("Expected type 'chat', got %s", prompt.Type)
	}

	if prompt.Name != "test-prompt" {
		t.Errorf("Expected name 'test-prompt', got %s", prompt.Name)
	}

	if prompt.Config != nil {
		t.Errorf("Expected config to be nil, got %v", prompt.Config)
	}

	if prompt.CommitMessage != "" {
		t.Errorf("Expected commitMessage to be empty, got %s", prompt.CommitMessage)
	}
}

func setupPromptsTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 1
	retryClient.RetryWaitMin = 1 * time.Millisecond
	retryClient.RetryWaitMax = 10 * time.Millisecond
	retryClient.Logger = nil

	client := &Client{
		retryableClient: retryClient,
		baseUrl:         server.URL,
		base64Token:     "test-token",
	}

	client.Prompts = (*PromptsService)(&service{client: client})

	return client, server
}

func TestPromptsService_GetPrompts_Success(t *testing.T) {
	expectedResponse := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"name":    "prompt1",
				"version": 1,
				"labels":  []string{"production"},
			},
			map[string]interface{}{
				"name":    "prompt2",
				"version": 2,
				"labels":  []string{"staging"},
			},
		},
		"meta": map[string]interface{}{
			"totalItems": 2,
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if r.URL.Path != "/api/public/v2/prompts" {
			t.Errorf("Expected path /api/public/v2/prompts, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	prompts, err := client.Prompts.GetPrompts()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if prompts == nil {
		t.Fatal("Expected prompts data, got nil")
	}

	// Verify data structure
	data, ok := prompts["data"].([]interface{})
	if !ok {
		t.Fatal("Expected 'data' field to be an array")
	}

	if len(data) != 2 {
		t.Errorf("Expected 2 prompts, got %d", len(data))
	}
}

func TestPromptsService_GetPrompts_Error(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	_, err := client.Prompts.GetPrompts()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestPromptsService_GetPromptByName_Success(t *testing.T) {
	expectedPrompt := Prompt{
		Name:    "test-prompt",
		Type:    "text",
		Version: 1,
		Labels:  []string{"production"},
		Tags:    []string{"test"},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if r.URL.Path != "/api/public/v2/prompts/test-prompt" {
			t.Errorf("Expected path /api/public/v2/prompts/test-prompt, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedPrompt)
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	prompt, err := client.Prompts.GetPromptByName("test-prompt", "", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if prompt == nil {
		t.Fatal("Expected prompt, got nil")
	}

	if prompt.Name != expectedPrompt.Name {
		t.Errorf("Expected name %s, got %s", expectedPrompt.Name, prompt.Name)
	}

	if prompt.Type != expectedPrompt.Type {
		t.Errorf("Expected type %s, got %s", expectedPrompt.Type, prompt.Type)
	}

	if prompt.Version != expectedPrompt.Version {
		t.Errorf("Expected version %d, got %d", expectedPrompt.Version, prompt.Version)
	}
}

func TestPromptsService_GetPromptByName_WithLabelAndVersion(t *testing.T) {
	version := 2

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		query := r.URL.Query()
		if query.Get("label") != "production" {
			t.Errorf("Expected label=production, got %s", query.Get("label"))
		}

		if query.Get("version") != "2" {
			t.Errorf("Expected version=2, got %s", query.Get("version"))
		}

		prompt := Prompt{
			Name:    "test-prompt",
			Type:    "chat",
			Version: version,
			Labels:  []string{"production"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(prompt)
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	prompt, err := client.Prompts.GetPromptByName("test-prompt", "production", &version)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if prompt.Version != version {
		t.Errorf("Expected version %d, got %d", version, prompt.Version)
	}
}

func TestPromptsService_GetPromptByName_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("prompt not found"))
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	_, err := client.Prompts.GetPromptByName("nonexistent", "", nil)
	if err == nil {
		t.Fatal("Expected error for not found prompt, got nil")
	}
}

func TestPromptsService_CreatePrompt_Success(t *testing.T) {
	newPrompt := &Prompt{
		Type: "chat",
		Name: "test-chat-prompt",
		Prompt: []ChatMessage{
			{
				Type:    "chatmessage",
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
		},
		Labels:        []string{"production", "v1"},
		Tags:          []string{"chat", "test"},
		CommitMessage: "Initial version",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify path
		if r.URL.Path != "/api/public/v2/prompts" {
			t.Errorf("Expected path /api/public/v2/prompts, got %s", r.URL.Path)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Decode and verify request body
		var receivedPrompt Prompt
		if err := json.NewDecoder(r.Body).Decode(&receivedPrompt); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if receivedPrompt.Name != newPrompt.Name {
			t.Errorf("Expected name %s, got %s", newPrompt.Name, receivedPrompt.Name)
		}

		if receivedPrompt.Type != newPrompt.Type {
			t.Errorf("Expected type %s, got %s", newPrompt.Type, receivedPrompt.Type)
		}

		if receivedPrompt.CommitMessage != newPrompt.CommitMessage {
			t.Errorf("Expected commit message %s, got %s", newPrompt.CommitMessage, receivedPrompt.CommitMessage)
		}

		// Return created prompt with version
		createdPrompt := receivedPrompt
		createdPrompt.Version = 1

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdPrompt)
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	createdPrompt, err := client.Prompts.CreatePrompt(newPrompt)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if createdPrompt == nil {
		t.Fatal("Expected created prompt, got nil")
	}

	if createdPrompt.Name != newPrompt.Name {
		t.Errorf("Expected name %s, got %s", newPrompt.Name, createdPrompt.Name)
	}

	if createdPrompt.Version != 1 {
		t.Errorf("Expected version 1, got %d", createdPrompt.Version)
	}
}

func TestPromptsService_CreatePrompt_ValidationError(t *testing.T) {
	newPrompt := &Prompt{
		Type: "chat",
		Name: "", // Invalid: empty name
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "name is required"}`))
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	_, err := client.Prompts.CreatePrompt(newPrompt)
	if err == nil {
		t.Fatal("Expected error for invalid prompt, got nil")
	}
}

func TestPromptsService_CreatePrompt_ServerError(t *testing.T) {
	newPrompt := &Prompt{
		Type: "chat",
		Name: "test-prompt",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	_, err := client.Prompts.CreatePrompt(newPrompt)
	if err == nil {
		t.Fatal("Expected error for server error, got nil")
	}
}

func TestPromptsService_CreatePrompt_TextPrompt(t *testing.T) {
	newPrompt := &Prompt{
		Type:          "text",
		Name:          "simple-text-prompt",
		Prompt:        "This is a simple text prompt",
		Labels:        []string{"test"},
		Tags:          []string{"simple"},
		CommitMessage: "Initial text prompt",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var receivedPrompt Prompt
		if err := json.NewDecoder(r.Body).Decode(&receivedPrompt); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify it's a text prompt
		if receivedPrompt.Type != "text" {
			t.Errorf("Expected type 'text', got %s", receivedPrompt.Type)
		}

		promptStr, ok := receivedPrompt.Prompt.(string)
		if !ok {
			t.Error("Expected prompt to be a string")
		} else if promptStr != "This is a simple text prompt" {
			t.Errorf("Expected prompt text, got %s", promptStr)
		}

		createdPrompt := receivedPrompt
		createdPrompt.Version = 1

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdPrompt)
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	createdPrompt, err := client.Prompts.CreatePrompt(newPrompt)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if createdPrompt.Type != "text" {
		t.Errorf("Expected type 'text', got %s", createdPrompt.Type)
	}
}

func TestPromptsService_UpdatePromptVersionLabels_Success(t *testing.T) {
	promptName := "test-prompt"
	version := 1
	newLabels := []string{"staging", "beta"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH method, got %s", r.Method)
		}

		// Verify path
		expectedPath := fmt.Sprintf("/api/public/v2/prompts/%s/versions/%d", promptName, version)
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Decode and verify request body
		var request UpdatePromptVersionLabelsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(request.NewLabels) != len(newLabels) {
			t.Errorf("Expected %d labels, got %d", len(newLabels), len(request.NewLabels))
		}

		for i, label := range newLabels {
			if request.NewLabels[i] != label {
				t.Errorf("Expected label %s at index %d, got %s", label, i, request.NewLabels[i])
			}
		}

		// Return updated prompt
		updatedPrompt := Prompt{
			Name:    promptName,
			Type:    "chat",
			Version: version,
			Labels:  newLabels,
			Tags:    []string{"test"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedPrompt)
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	updatedPrompt, err := client.Prompts.UpdatePromptVersionLabels(promptName, version, newLabels)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedPrompt == nil {
		t.Fatal("Expected updated prompt, got nil")
	}

	if updatedPrompt.Name != promptName {
		t.Errorf("Expected name %s, got %s", promptName, updatedPrompt.Name)
	}

	if updatedPrompt.Version != version {
		t.Errorf("Expected version %d, got %d", version, updatedPrompt.Version)
	}

	if len(updatedPrompt.Labels) != len(newLabels) {
		t.Errorf("Expected %d labels, got %d", len(newLabels), len(updatedPrompt.Labels))
	}

	for i, label := range newLabels {
		if updatedPrompt.Labels[i] != label {
			t.Errorf("Expected label %s at index %d, got %s", label, i, updatedPrompt.Labels[i])
		}
	}
}

func TestPromptsService_UpdatePromptVersionLabels_ValidationError(t *testing.T) {
	promptName := "test-prompt"
	version := 1
	newLabels := []string{"latest"} // 'latest' is a reserved label

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "The 'latest' label is reserved and managed by Langfuse"}`))
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	_, err := client.Prompts.UpdatePromptVersionLabels(promptName, version, newLabels)
	if err == nil {
		t.Fatal("Expected error for reserved label, got nil")
	}
}

func TestPromptsService_UpdatePromptVersionLabels_DuplicateLabels(t *testing.T) {
	promptName := "test-prompt"
	version := 1
	newLabels := []string{"production"} // Label already exists on another version

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Labels must be unique across versions"}`))
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	_, err := client.Prompts.UpdatePromptVersionLabels(promptName, version, newLabels)
	if err == nil {
		t.Fatal("Expected error for duplicate label, got nil")
	}
}

func TestPromptsService_UpdatePromptVersionLabels_NotFound(t *testing.T) {
	promptName := "nonexistent-prompt"
	version := 99
	newLabels := []string{"test"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Prompt version not found"}`))
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	_, err := client.Prompts.UpdatePromptVersionLabels(promptName, version, newLabels)
	if err == nil {
		t.Fatal("Expected error for not found prompt version, got nil")
	}
}

func TestPromptsService_UpdatePromptVersionLabels_Unauthorized(t *testing.T) {
	promptName := "test-prompt"
	version := 1
	newLabels := []string{"staging"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	_, err := client.Prompts.UpdatePromptVersionLabels(promptName, version, newLabels)
	if err == nil {
		t.Fatal("Expected error for unauthorized request, got nil")
	}
}

func TestPromptsService_UpdatePromptVersionLabels_EmptyLabels(t *testing.T) {
	promptName := "test-prompt"
	version := 1
	newLabels := []string{} // Empty labels array

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		var request UpdatePromptVersionLabelsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if len(request.NewLabels) != 0 {
			t.Errorf("Expected empty labels array, got %d labels", len(request.NewLabels))
		}

		// Return updated prompt with empty labels
		updatedPrompt := Prompt{
			Name:    promptName,
			Type:    "chat",
			Version: version,
			Labels:  []string{},
			Tags:    []string{"test"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedPrompt)
	}

	client, server := setupPromptsTestClient(handler)
	defer server.Close()

	updatedPrompt, err := client.Prompts.UpdatePromptVersionLabels(promptName, version, newLabels)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(updatedPrompt.Labels) != 0 {
		t.Errorf("Expected 0 labels, got %d", len(updatedPrompt.Labels))
	}
}

// TestPromptsService_GetPromptByName_WithSpecialCharactersInName tests URL encoding for special characters in prompt name
func TestPromptsService_GetPromptByName_WithSpecialCharactersInName(t *testing.T) {
	testCases := []struct {
		name         string
		promptName   string
		expectedPath string
	}{
		{
			name:         "prompt name with spaces",
			promptName:   "my prompt name",
			expectedPath: "/api/public/v2/prompts/my%20prompt%20name",
		},
		{
			name:         "prompt name with forward slash",
			promptName:   "path/to/prompt",
			expectedPath: "/api/public/v2/prompts/path%2Fto%2Fprompt",
		},
		{
			name:         "prompt name with special characters",
			promptName:   "prompt@test#123",
			expectedPath: "/api/public/v2/prompts/prompt@test%23123",
		},
		{
			name:         "prompt name with question mark",
			promptName:   "what?is?this",
			expectedPath: "/api/public/v2/prompts/what%3Fis%3Fthis",
		},
		{
			name:         "prompt name with ampersand",
			promptName:   "this&that",
			expectedPath: "/api/public/v2/prompts/this&that",
		},
		{
			name:         "prompt name with percent sign",
			promptName:   "100%complete",
			expectedPath: "/api/public/v2/prompts/100%25complete",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				// Verify the URL path is correctly encoded by checking EscapedPath
				if r.URL.EscapedPath() != tc.expectedPath {
					t.Errorf("Expected path %s, got %s", tc.expectedPath, r.URL.EscapedPath())
				}

				prompt := Prompt{
					Name:    tc.promptName,
					Type:    "text",
					Version: 1,
					Labels:  []string{"production"},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(prompt)
			}

			client, server := setupPromptsTestClient(handler)
			defer server.Close()

			prompt, err := client.Prompts.GetPromptByName(tc.promptName, "", nil)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if prompt.Name != tc.promptName {
				t.Errorf("Expected name %s, got %s", tc.promptName, prompt.Name)
			}
		})
	}
}

// TestPromptsService_GetPromptByName_WithSpecialCharactersInLabel tests URL encoding for special characters in label
func TestPromptsService_GetPromptByName_WithSpecialCharactersInLabel(t *testing.T) {
	testCases := []struct {
		name          string
		label         string
		expectedLabel string
	}{
		{
			name:          "label with spaces",
			label:         "my label",
			expectedLabel: "my+label",
		},
		{
			name:          "label with special characters",
			label:         "v1.0-beta",
			expectedLabel: "v1.0-beta",
		},
		{
			name:          "label with ampersand",
			label:         "test&prod",
			expectedLabel: "test%26prod",
		},
		{
			name:          "label with equals sign",
			label:         "version=1",
			expectedLabel: "version%3D1",
		},
		{
			name:          "label with percent sign",
			label:         "100%",
			expectedLabel: "100%25",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				// Verify query parameter is correctly encoded
				query := r.URL.Query()
				receivedLabel := query.Get("label")

				// The expected label should match what we sent
				if receivedLabel != tc.label {
					t.Errorf("Expected label query param to decode to %s, got %s", tc.label, receivedLabel)
				}

				// Check the raw query string contains the encoded value
				if !contains(r.URL.RawQuery, "label="+tc.expectedLabel) {
					t.Errorf("Expected raw query to contain 'label=%s', got %s", tc.expectedLabel, r.URL.RawQuery)
				}

				prompt := Prompt{
					Name:    "test-prompt",
					Type:    "text",
					Version: 1,
					Labels:  []string{tc.label},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(prompt)
			}

			client, server := setupPromptsTestClient(handler)
			defer server.Close()

			prompt, err := client.Prompts.GetPromptByName("test-prompt", tc.label, nil)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(prompt.Labels) > 0 && prompt.Labels[0] != tc.label {
				t.Errorf("Expected label %s, got %s", tc.label, prompt.Labels[0])
			}
		})
	}
}

// TestPromptsService_UpdatePromptVersionLabels_WithSpecialCharactersInName tests URL encoding for special characters in prompt name
func TestPromptsService_UpdatePromptVersionLabels_WithSpecialCharactersInName(t *testing.T) {
	testCases := []struct {
		name         string
		promptName   string
		expectedPath string
	}{
		{
			name:         "prompt name with spaces",
			promptName:   "my prompt",
			expectedPath: "/api/public/v2/prompts/my%20prompt/versions/1",
		},
		{
			name:         "prompt name with forward slash",
			promptName:   "org/prompt",
			expectedPath: "/api/public/v2/prompts/org%2Fprompt/versions/1",
		},
		{
			name:         "prompt name with special characters",
			promptName:   "prompt@v1#test",
			expectedPath: "/api/public/v2/prompts/prompt@v1%23test/versions/1",
		},
		{
			name:         "prompt name with percent sign",
			promptName:   "100%done",
			expectedPath: "/api/public/v2/prompts/100%25done/versions/1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			version := 1
			newLabels := []string{"staging"}

			handler := func(w http.ResponseWriter, r *http.Request) {
				// Verify the URL path is correctly encoded by checking EscapedPath
				if r.URL.EscapedPath() != tc.expectedPath {
					t.Errorf("Expected path %s, got %s", tc.expectedPath, r.URL.EscapedPath())
				}

				updatedPrompt := Prompt{
					Name:    tc.promptName,
					Type:    "chat",
					Version: version,
					Labels:  newLabels,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(updatedPrompt)
			}

			client, server := setupPromptsTestClient(handler)
			defer server.Close()

			prompt, err := client.Prompts.UpdatePromptVersionLabels(tc.promptName, version, newLabels)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if prompt.Name != tc.promptName {
				t.Errorf("Expected name %s, got %s", tc.promptName, prompt.Name)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
