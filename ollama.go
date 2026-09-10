package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ollama/ollama/api"
)

// ollamaGenerator talks to an Ollama server via /api/generate.
type ollamaGenerator struct {
	client *api.Client
}

// createOllamaClient initializes an Ollama client with the specified server and timeout.
func createOllamaClient(serverURL string, timeoutSecs int) (*api.Client, error) {
	httpClient := newHTTPClient(timeoutSecs)

	baseURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL '%s': %w", serverURL, err)
	}

	return api.NewClient(baseURL, httpClient), nil
}

// newOllamaGenerator creates a generator backed by Ollama.
func newOllamaGenerator(serverURL string, timeoutSecs int) (generator, error) {
	client, err := createOllamaClient(serverURL, timeoutSecs)
	if err != nil {
		return nil, err
	}
	return &ollamaGenerator{client: client}, nil
}

// Generate sends a generate request to Ollama and returns the full response text.
func (g *ollamaGenerator) Generate(ctx context.Context, model, prompt string, image []byte, mediaType string, seed int) (string, error) {
	var images []api.ImageData
	if image != nil {
		images = []api.ImageData{image}
	}

	req := &api.GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Images: images,
		Options: map[string]interface{}{
			"seed": seed,
		},
	}

	var response strings.Builder
	respFunc := func(resp api.GenerateResponse) error {
		response.WriteString(resp.Response)
		return nil
	}

	if err := g.client.Generate(ctx, req, respFunc); err != nil {
		return "", err
	}

	return response.String(), nil
}
