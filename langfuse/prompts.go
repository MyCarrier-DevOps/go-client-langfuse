package langfuse

// PromptsService handles operations related to prompts
type PromptsService service

// Prompt represents a prompt in langfuse
type Prompt struct {
	Config        map[string]interface{}
	CommitMessage string
	Labels        map[string]string
	Name          string
	Version       int
	Tags          map[string]string
	Type          string
}
