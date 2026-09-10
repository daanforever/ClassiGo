package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxErrorBodyLen = 512

// openaiGenerator talks to an OpenAI-compatible chat completions endpoint.
type openaiGenerator struct {
	httpClient *http.Client
	baseURL    string
}

type openaiChatRequest struct {
	Model    string              `json:"model"`
	Messages []openaiChatMessage `json:"messages"`
	Seed     int                 `json:"seed"`
	Stream   bool                `json:"stream"`
}

type openaiChatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []openaiContentPart
}

type openaiContentPart struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *openaiImageURL `json:"image_url,omitempty"`
}

type openaiImageURL struct {
	URL string `json:"url"`
}

type openaiChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// chatCompletionsURL builds the chat completions endpoint from a server origin.
func chatCompletionsURL(serverURL string) string {
	return strings.TrimRight(serverURL, "/") + "/v1/chat/completions"
}

// newOpenAIGenerator creates a generator for OpenAI-compatible APIs.
func newOpenAIGenerator(serverURL string, timeoutSecs int) (generator, error) {
	return &openaiGenerator{
		httpClient: newHTTPClient(timeoutSecs),
		baseURL:    serverURL,
	}, nil
}

// buildOpenAIChatRequest builds the chat completions request body.
func buildOpenAIChatRequest(model, prompt string, image []byte, mediaType string, seed int) openaiChatRequest {
	msg := openaiChatMessage{Role: "user"}
	if image == nil {
		msg.Content = prompt
	} else {
		dataURL := fmt.Sprintf("data:%s;base64,%s", mediaType, base64.StdEncoding.EncodeToString(image))
		msg.Content = []openaiContentPart{
			{Type: "text", Text: prompt},
			{Type: "image_url", ImageURL: &openaiImageURL{URL: dataURL}},
		}
	}
	return openaiChatRequest{
		Model:    model,
		Messages: []openaiChatMessage{msg},
		Seed:     seed,
		Stream:   false,
	}
}

// Generate sends a chat completions request and returns the assistant message content.
func (g *openaiGenerator) Generate(ctx context.Context, model, prompt string, image []byte, mediaType string, seed int) (string, error) {
	reqBody := buildOpenAIChatRequest(model, prompt, image, mediaType, seed)

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := chatCompletionsURL(g.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody := string(body)
		if len(errBody) > maxErrorBodyLen {
			errBody = errBody[:maxErrorBodyLen] + "..."
		}
		return "", fmt.Errorf("chat completions returned status %d: %s", resp.StatusCode, errBody)
	}

	var chatResp openaiChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", nil
	}

	return chatResp.Choices[0].Message.Content, nil
}
