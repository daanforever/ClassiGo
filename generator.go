package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// generator abstracts LLM backends (Ollama, OpenAI-compatible, etc.).
type generator interface {
	Generate(ctx context.Context, model, prompt string, image []byte, mediaType string, seed int) (string, error)
}

// validateAPI normalizes and validates the --api flag value.
func validateAPI(s string) (string, error) {
	api := strings.ToLower(strings.TrimSpace(s))
	switch api {
	case "ollama", "openai":
		return api, nil
	default:
		return "", fmt.Errorf("unsupported --api value %q (allowed: ollama, openai)", s)
	}
}

// createGenerator returns an LLM generator for the configured API.
func createGenerator(config *Config) (generator, error) {
	switch config.API {
	case "openai":
		return newOpenAIGenerator(config.ServerURL, config.Timeout)
	default: // ollama
		return newOllamaGenerator(config.ServerURL, config.Timeout)
	}
}

// newHTTPClient creates an HTTP client with optional timeout.
func newHTTPClient(timeoutSecs int) *http.Client {
	httpClient := &http.Client{}
	if timeoutSecs > 0 {
		httpClient.Timeout = time.Duration(timeoutSecs) * time.Second
	}
	return httpClient
}

// imageMediaType returns a MIME type for a file extension.
func imageMediaType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
