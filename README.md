# go-client-langfuse

[![Go Reference](https://pkg.go.dev/badge/github.com/MyCarrier-DevOps/go-client-langfuse.svg)](https://pkg.go.dev/github.com/MyCarrier-DevOps/go-client-langfuse)
[![Go Report Card](https://goreportcard.com/badge/github.com/MyCarrier-DevOps/go-client-langfuse)](https://goreportcard.com/report/github.com/MyCarrier-DevOps/go-client-langfuse)
[![CI Status](https://github.com/MyCarrier-DevOps/go-client-langfuse/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/MyCarrier-DevOps/go-client-langfuse/actions/workflows/ci.yml)

<i> ⚠️ Note: This library is not officially maintained by Langfuse. It’s a community-driven project and is actively being developed. </i>


go-client-langfuse is a Go client library for accessing the [Langfuse API](https://api.reference.langfuse.com).

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [Authentication](#authentication)
  - [Creating a Client](#creating-a-client)
  - [Projects](#projects)
  - [Prompts](#prompts)
- [Examples](#examples)
- [Error Handling](#error-handling)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Installation

go-client-langfuse is compatible with modern Go releases in module mode, with Go installed:

```bash
go get github.com/MyCarrier-DevOps/go-client-langfuse@v1.0.0
```

will resolve and add the package to the current development module, along with its dependencies.

Alternatively the same can be achieved if you use import in a package:

```go
import "github.com/MyCarrier-DevOps/go-client-langfuse/langfuse"
```

and run `go get` without parameters.

## Usage

```go
import "github.com/MyCarrier-DevOps/go-client-langfuse/langfuse"
```

### Authentication

The Langfuse API uses HTTP Basic Authentication with your public and secret keys. The library handles this automatically using environment variables.

Set the following environment variables:

```bash
export LANGFUSE_SERVER_URL="https://cloud.langfuse.com"  # or your self-hosted URL
export LANGFUSE_PUBLIC_KEY="pk-lf-..."
export LANGFUSE_SECRET_KEY="sk-lf-..."
```

### Creating a Client

You can create a client in two ways:

#### Option 1: Using Environment Variables

Load the configuration from environment variables, then create a client:

```go
package main

import (
    "log"
    "github.com/MyCarrier-DevOps/go-client-langfuse/langfuse"
)

func main() {
    // Load configuration from environment variables
    config, err := langfuse.LoadConfigFromEnvVars()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Create a new client with the loaded configuration
    client := langfuse.NewClient(config)

    // Use the client...
}
```

#### Option 2: Direct Configuration (Without Environment Variables)

Create a configuration directly without environment variables:

```go
package main

import (
    "log"
    "github.com/MyCarrier-DevOps/go-client-langfuse/langfuse"
)

func main() {
    // Create configuration directly
    config, err := langfuse.NewConfig(
        "https://cloud.langfuse.com",
        "pk-lf-xxx",
        "sk-lf-xxx",
    )
    if err != nil {
        log.Fatalf("Failed to create config: %v", err)
    }

    // Create a new client with the config
    client := langfuse.NewClient(config)

    // Use the client...
}
```

The client uses retryable HTTP requests with the following default configuration:
- **Max Retries**: 3 attempts
- **Retry Wait Min**: 1 second
- **Retry Wait Max**: 4 seconds
- **Retry Policy**: Retries on 5xx errors and network failures

### Projects

Get information about the project associated with your API keys:

```go
project, err := client.Projects.GetProject()
if err != nil {
    log.Fatalf("Error fetching project: %v", err)
}

fmt.Printf("Project ID: %v\n", project["id"])
fmt.Printf("Project Name: %v\n", project["name"])
```

### Prompts

The library provides comprehensive support for managing prompts in Langfuse.

#### List All Prompts

```go
prompts, err := client.Prompts.GetPrompts()
if err != nil {
    log.Fatalf("Error fetching prompts: %v", err)
}

fmt.Printf("Prompts: %+v\n", prompts)
```

#### Get a Specific Prompt

Retrieve a prompt by name with optional label or version:

```go
// Get latest version with no label
prompt, err := client.Prompts.GetPromptByName("my-prompt", "", nil)
if err != nil {
    log.Fatalf("Error fetching prompt: %v", err)
}

// Get prompt with specific label
prompt, err := client.Prompts.GetPromptByName("my-prompt", "production", nil)

// Get specific version
version := 3
prompt, err := client.Prompts.GetPromptByName("my-prompt", "", &version)
```

#### Create a New Prompt

Create a text prompt:

```go
newPrompt := &langfuse.Prompt{
    Name:          "simple-text-prompt",
    Type:          "text",
    Prompt:        "Translate the following text to French: {{text}}",
    Labels:        []string{"production"},
    Tags:          []string{"translation", "french"},
    CommitMessage: "Initial version",
}

createdPrompt, err := client.Prompts.CreatePrompt(newPrompt)
if err != nil {
    log.Fatalf("Error creating prompt: %v", err)
}
```

Create a chat prompt:

```go
chatPrompt := &langfuse.Prompt{
    Name: "chat-assistant",
    Type: "chat",
    Prompt: []langfuse.ChatMessage{
        {
            Type:    "chatmessage",
            Role:    "system",
            Content: "You are a helpful assistant that answers questions concisely.",
        },
        {
            Type:    "chatmessage",
            Role:    "user",
            Content: "{{user_question}}",
        },
    },
    Labels:        []string{"production", "v1"},
    Tags:          []string{"chat", "assistant"},
    CommitMessage: "Initial chat prompt",
}

createdPrompt, err := client.Prompts.CreatePrompt(chatPrompt)
if err != nil {
    log.Fatalf("Error creating chat prompt: %v", err)
}
```

#### Update Prompt Version Labels

Update the labels for a specific prompt version. Note that labels must be unique across all versions of a prompt, and the `latest` label is reserved and managed by Langfuse:

```go
// Update labels for version 1 of "my-prompt"
updatedPrompt, err := client.Prompts.UpdatePromptVersionLabels(
    "my-prompt",
    1,
    []string{"staging", "beta"},
)
if err != nil {
    log.Fatalf("Error updating prompt version labels: %v", err)
}

fmt.Printf("Updated prompt version %d with labels: %v\n",
    updatedPrompt.Version,
    updatedPrompt.Labels)
```

Common use cases:
- Promote a version from "staging" to "production"
- Add experiment labels like "beta" or "canary"
- Remove labels by providing a new list that excludes them
- Clear all labels by providing an empty slice `[]string{}`

## Examples

For a complete working example, see [example/example.go](example/example.go).

To run the example:

```bash
# Set environment variables
export LANGFUSE_SERVER_URL="https://cloud.langfuse.com"
export LANGFUSE_PUBLIC_KEY="your-public-key"
export LANGFUSE_SECRET_KEY="your-secret-key"

# Run the example
go run example/example.go
```

## Error Handling

The library uses Go's standard error handling. All API methods return an error as the last return value:

```go
project, err := client.Projects.GetProject()
if err != nil {
    // Handle error
    log.Printf("Error: %v", err)
    return
}
```

Errors are wrapped with context to help identify where issues occurred:

- `error making request:` - HTTP request failed (network issues, timeouts)
- `client error 4xx:` - Client errors (bad request, unauthorized, not found, etc.)
- `server error 5xx:` - Server errors (internal server error, service unavailable)
- `error unmarshalling:` - Failed to parse JSON response
- `error fetching project:` - Project-specific operation failed
- `error fetching prompt:` - Prompt-specific operation failed

## Testing

The library includes comprehensive test coverage. To run tests:

```bash
# Run all tests
go test ./langfuse

# Run with verbose output
go test -v ./langfuse

# Run with coverage
go test -cover ./langfuse

# Run specific test
go test -v -run TestClient_Do_Success ./langfuse
```

Tests use mock HTTP servers to avoid making real API calls.

## API Coverage

The library currently supports the following Langfuse API endpoints:

### Projects API
- `GET /api/public/projects` - Get project information

### Prompts API
- `GET /api/public/v2/prompts` - List all prompts
- `GET /api/public/v2/prompts/{name}` - Get prompt by name (with optional label/version)
- `POST /api/public/v2/prompts` - Create a new prompt or version
- `PATCH /api/public/v2/prompts/{name}/versions/{version}` - Update prompt version labels


## Roadmap

Future enhancements planned:
- Support for Traces API
- Support for Generations API
- Support for Observations API
- Support for Datasets API
- Support for Scores API
- Context-aware request cancellation
- Additional configuration options
- Pagination helpers

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

When contributing:
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Run linting (`golangci-lint run`)
6. Commit your changes (`git commit -m 'Add some amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Versioning

This library follows [Semantic Versioning](https://semver.org/).

## Resources

- [Langfuse Documentation](https://langfuse.com/docs)
- [Langfuse API Reference](https://api.reference.langfuse.com/)
- [Langfuse GitHub](https://github.com/langfuse/langfuse)


## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/MyCarrier-DevOps/go-client-langfuse).
