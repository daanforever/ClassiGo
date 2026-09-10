package main

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateAPI(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "ollama default", input: "ollama", want: "ollama"},
		{name: "openai", input: "openai", want: "openai"},
		{name: "uppercase", input: "OpenAI", want: "openai"},
		{name: "with spaces", input: "  OLLAMA  ", want: "ollama"},
		{name: "invalid", input: "anthropic", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateAPI(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestChatCompletionsURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "no trailing slash", input: "http://192.168.1.45:1234", want: "http://192.168.1.45:1234/v1/chat/completions"},
		{name: "trailing slash", input: "http://192.168.1.45:1234/", want: "http://192.168.1.45:1234/v1/chat/completions"},
		{name: "localhost", input: "http://localhost:1234", want: "http://localhost:1234/v1/chat/completions"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := chatCompletionsURL(tc.input)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestImageMediaType(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".JPG", "image/jpeg"},
		{".png", "image/png"},
		{".gif", "image/gif"},
		{".bmp", "image/bmp"},
		{".webp", "image/webp"},
		{".tiff", "application/octet-stream"},
		{"", "application/octet-stream"},
	}

	for _, tc := range tests {
		t.Run(tc.ext, func(t *testing.T) {
			got := imageMediaType(tc.ext)
			if got != tc.want {
				t.Errorf("imageMediaType(%q) = %q, want %q", tc.ext, got, tc.want)
			}
		})
	}
}

func TestBuildOpenAIChatRequestTextOnly(t *testing.T) {
	req := buildOpenAIChatRequest("my-model", "describe this", nil, "", 42)

	if req.Model != "my-model" {
		t.Errorf("model = %q, want my-model", req.Model)
	}
	if req.Seed != 42 {
		t.Errorf("seed = %d, want 42", req.Seed)
	}
	if req.Stream {
		t.Error("stream should be false")
	}
	if len(req.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(req.Messages))
	}
	if req.Messages[0].Role != "user" {
		t.Errorf("role = %q, want user", req.Messages[0].Role)
	}

	content, ok := req.Messages[0].Content.(string)
	if !ok {
		t.Fatalf("content type = %T, want string", req.Messages[0].Content)
	}
	if content != "describe this" {
		t.Errorf("content = %q, want describe this", content)
	}

	// Ensure seed is present in JSON even when non-zero (and not omitted)
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"seed":42`) {
		t.Errorf("marshaled JSON missing seed: %s", data)
	}
	if strings.Contains(string(data), "image_url") {
		t.Errorf("text-only request should not include image_url: %s", data)
	}
}

func TestBuildOpenAIChatRequestWithImage(t *testing.T) {
	image := []byte{0x01, 0x02, 0x03}
	req := buildOpenAIChatRequest("vision-model", "what is this?", image, "image/png", 7)

	parts, ok := req.Messages[0].Content.([]openaiContentPart)
	if !ok {
		t.Fatalf("content type = %T, want []openaiContentPart", req.Messages[0].Content)
	}
	if len(parts) != 2 {
		t.Fatalf("expected 2 content parts, got %d", len(parts))
	}
	if parts[0].Type != "text" || parts[0].Text != "what is this?" {
		t.Errorf("text part = %+v", parts[0])
	}
	if parts[1].Type != "image_url" || parts[1].ImageURL == nil {
		t.Fatalf("image part = %+v", parts[1])
	}

	wantPrefix := "data:image/png;base64,"
	if !strings.HasPrefix(parts[1].ImageURL.URL, wantPrefix) {
		t.Errorf("data URL = %q, want prefix %q", parts[1].ImageURL.URL, wantPrefix)
	}
	encoded := strings.TrimPrefix(parts[1].ImageURL.URL, wantPrefix)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode base64: %v", err)
	}
	if string(decoded) != string(image) {
		t.Errorf("decoded image = %v, want %v", decoded, image)
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"seed":7`) {
		t.Errorf("marshaled JSON missing seed: %s", data)
	}
	if !strings.Contains(string(data), "image_url") {
		t.Errorf("image request should include image_url: %s", data)
	}
}

func TestBuildOpenAIChatRequestSeedZero(t *testing.T) {
	req := buildOpenAIChatRequest("m", "p", nil, "", 0)
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"seed":0`) {
		t.Errorf("seed 0 must not be omitted: %s", data)
	}
}
