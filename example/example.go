package main

import (
	langfuse "github.com/MyCarrier-DevOps/go-client-langfuse/langfuse"
	"github.com/MyCarrier-DevOps/goLibMyCarrier/otel"
)

func main() {
	// Initialize OpenTelemetry (optional, but recommended for tracing)
	log := otel.NewAppLogger()
	// Load configuration from environment variables
	if err := langfuse.LoadConfig(); err != nil {
		log.Errorf("Failed to load config: %v", err)
		return
	}

	// Create a new client using the loaded config
	client := langfuse.NewClient()

	// Use the client to get project information
	project, err := client.Projects.GetProject()
	if err != nil {
		log.Errorf("Failed to get project: %v", err)
		return
	}

	// Log project information
	log.Infof("Project info: %+v", project)

	// Example: Create a new chat prompt
	newPrompt := &langfuse.Prompt{
		Type: "chat",
		Name: "example-chat-prompt",
		Prompt: []langfuse.ChatMessage{
			{
				Type:    "chatmessage",
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Type:    "chatmessage",
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		Labels:        []string{"production", "v1"},
		Tags:          []string{"chat", "example"},
		CommitMessage: "Initial version of example chat prompt",
	}

	createdPrompt, err := client.Prompts.CreatePrompt(newPrompt)
	if err != nil {
		log.Errorf("Failed to create prompt: %v", err)
		return
	}

	// Log created prompt information
	log.Infof("Created Prompt: %+v", createdPrompt)

	// Example: Retrieve the created prompt by name
	retrievedPrompt, err := client.Prompts.GetPromptByName(createdPrompt.Name, "", nil)
	if err != nil {
		log.Errorf("Failed to retrieve prompt: %v", err)
		return
	}

	// Log retrieved prompt information
	log.Infof("Retrieved Prompt: %+v", retrievedPrompt)

	// Example: List all prompts
	prompts, err := client.Prompts.GetPrompts()
	if err != nil {
		log.Errorf("Failed to list prompts: %v", err)
		return
	}

	// Log list of prompts
	log.Infof("List of Prompts: %+v", prompts)
}
