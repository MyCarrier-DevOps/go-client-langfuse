package langfuse

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// PromptsService handles operations related to prompts
type PromptsService service

// Prompt represents a prompt in langfuse
type Prompt struct {
	Config        map[string]interface{} `json:"config,omitempty"`
	CommitMessage string                 `json:"commitMessage,omitempty"`
	Labels        []string               `json:"labels,omitempty"`
	Name          string                 `json:"name"`
	Prompt        interface{}            `json:"prompt,omitempty"`
	Version       int                    `json:"version,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Type          string                 `json:"type"`
}

// ChatMessage represents a chat message in a chat prompt
type ChatMessage struct {
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

// UpdatePromptVersionLabelsRequest represents the request body for updating prompt version labels
type UpdatePromptVersionLabelsRequest struct {
	NewLabels []string `json:"newLabels"`
}

// Get a list of prompt names with versions and labels for the given API token
// https://api.reference.langfuse.com/#tag/prompts/get/api/public/v2/prompts
func (s *PromptsService) GetPrompts() (map[string]interface{}, error) {
	u := "/api/public/v2/prompts"

	body, err := s.client.Do("GET", u)
	if err != nil {
		return nil, fmt.Errorf("error fetching prompts: %w", err)
	}

	var promptsData map[string]interface{}
	err = json.Unmarshal(body, &promptsData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling prompts data: %w", err)
	}

	return promptsData, nil
}

// GetPromptByName retrieves a specific prompt by its Name
// https://api.reference.langfuse.com/#tag/prompts/get/api/public/v2/prompts/{promptName}
func (s *PromptsService) GetPromptByName(name, label string, version *int) (*Prompt, error) {
	// Build URL path with properly escaped name
	u := fmt.Sprintf("/api/public/v2/prompts/%s", url.PathEscape(name))

	// Build query parameters using url.Values for proper encoding
	queryParams := url.Values{}
	if label != "" {
		queryParams.Set("label", label)
	}
	if version != nil {
		queryParams.Set("version", fmt.Sprintf("%d", *version))
	}

	// Append query string if there are parameters
	if len(queryParams) > 0 {
		u = u + "?" + queryParams.Encode()
	}

	body, err := s.client.Do("GET", u)
	if err != nil {
		return nil, fmt.Errorf("error fetching prompt by name: %w", err)
	}

	var prompt Prompt
	err = json.Unmarshal(body, &prompt)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling prompt data: %w", err)
	}

	return &prompt, nil
}

// CreatePrompt creates a new prompt or a new version for an existing prompt
// https://api.reference.langfuse.com/#tag/prompts/post/api/public/v2/prompts
func (s *PromptsService) CreatePrompt(prompt *Prompt) (*Prompt, error) {
	u := "/api/public/v2/prompts"

	body, err := s.client.DoWithBody("POST", u, prompt)
	if err != nil {
		return nil, fmt.Errorf("error creating prompt: %w", err)
	}

	var createdPrompt Prompt
	err = json.Unmarshal(body, &createdPrompt)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling created prompt data: %w", err)
	}

	return &createdPrompt, nil
}

// UpdatePromptVersionLabels updates the labels for a specific prompt version
// https://api.reference.langfuse.com/#tag/promptversion/patch/api/public/v2/prompts/%7Bname%7D/versions/%7Bversion%7D
func (s *PromptsService) UpdatePromptVersionLabels(name string, version int, newLabels []string) (*Prompt, error) {
	// url encode name
	encodedName := url.PathEscape(name)
	u := fmt.Sprintf("/api/public/v2/prompts/%s/versions/%d", encodedName, version)

	request := &UpdatePromptVersionLabelsRequest{
		NewLabels: newLabels,
	}

	body, err := s.client.DoWithBody("PATCH", u, request)
	if err != nil {
		return nil, fmt.Errorf("error updating prompt version labels: %w", err)
	}

	var updatedPrompt Prompt
	err = json.Unmarshal(body, &updatedPrompt)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling updated prompt data: %w", err)
	}

	return &updatedPrompt, nil
}
