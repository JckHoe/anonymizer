# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based text anonymization service that uses OpenAI's API (via OpenRouter) to anonymize personal information in text. The service reads a system prompt template and processes user input to replace names, locations, and identifying information with generic placeholders.

## Architecture

The application follows a simple single-binary architecture:

- **cmd/main.go**: Entry point that sets up the OpenAI client, reads the system prompt template, and makes API calls
- **system_prompt.tmpl**: Template file containing the system prompt that instructs the AI to anonymize text and return JSON responses
- **go.mod**: Go module definition with OpenAI Go SDK dependency

The application uses the OpenRouter API (https://openrouter.ai/api/v1) as a proxy to OpenAI's services, requiring an `OPENROUTER_API_KEY` environment variable.

## Development Commands

### Running the Application
```bash
make run
# or
go run cmd/main.go
```

### Building
```bash
go build -o anonymizer cmd/main.go
```

### Dependencies
```bash
go mod tidy
go mod download
```

## Configuration

The application supports both local Ollama and OpenRouter/OpenAI:

### For Ollama (default):
- `OLLAMA_BASE_URL` (optional): Ollama server URL (defaults to `http://localhost:11434/v1`)
- `MODEL_NAME` (optional): Model name to use (defaults to `llama3.2`)
- `system_prompt.tmpl` file in the working directory

### For OpenRouter:
- `OPENROUTER_API_KEY`: Required for OpenRouter authentication
- `OLLAMA_BASE_URL`: Set to OpenRouter URL or leave unset to use OpenRouter
- `MODEL_NAME` (optional): Model name (defaults to `openai/gpt-4.1-mini` when using OpenRouter)
- `system_prompt.tmpl` file in the working directory

The application automatically detects whether to use Ollama or OpenRouter based on the presence of `OPENROUTER_API_KEY`.

## Key Implementation Details

- Uses OpenAI Go SDK v1.10.1 with configurable base URL (supports both Ollama and OpenRouter)
- Automatically selects appropriate model based on configuration
- Expects JSON responses with "anonymized_text" and "replacements" fields
- System prompt instructs the AI to track all replacements made during anonymization
- Currently has a hardcoded example message about "Amy" for demonstration purposes

## File Structure

```
.
├── cmd/
│   └── main.go          # Main application entry point
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── Makefile             # Build commands
└── system_prompt.tmpl   # AI system prompt template
```

## Testing

No test files are currently present in the repository. Consider adding tests for:
- System prompt template loading
- API client configuration
- Response parsing and validation