package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GroqProvider struct {
	APIKey    string
	ModelName string // default: llama-3.2-11b-vision-preview
}

func NewGroqProvider(apiKey, modelName string) *GroqProvider {
	if modelName == "" {
		modelName = "llama-3.2-11b-vision-preview"
	}
	return &GroqProvider{APIKey: apiKey, ModelName: modelName}
}

func (p *GroqProvider) Name() string       { return "groq" }
func (p *GroqProvider) Model() string      { return p.ModelName }
func (p *GroqProvider) IsConfigured() bool { return p.APIKey != "" }

// API Request/Response Types
type groqImageURL struct {
	URL string `json:"url"`
}

type groqContentItem struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageURL *groqImageURL `json:"image_url,omitempty"`
}

type groqMessage struct {
	Role    string            `json:"role"`
	Content []groqContentItem `json:"content"`
}

type groqResponseFormat struct {
	Type string `json:"type"`
}

type groqRequestPayload struct {
	Model          string             `json:"model"`
	Messages       []groqMessage      `json:"messages"`
	ResponseFormat groqResponseFormat `json:"response_format"`
}

type groqResponsePayload struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *GroqProvider) ExtractKTP(ctx context.Context, imageData []byte) (*KTPData, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("API Key Groq belum diatur")
	}

	b64Data := base64.StdEncoding.EncodeToString(imageData)

	prompt := `Ekstrak data dari gambar KTP-el berikut ke format JSON.
Kembalikan objek JSON dengan field: nik, nama, tempat_lahir, tanggal_lahir (YYYY-MM-DD), jenis_kelamin (L/P), alamat, rt, rw, kelurahan, kecamatan, agama, status_kawin, pekerjaan, kewarganegaraan, confidence.`

	payload := groqRequestPayload{
		Model: p.ModelName,
		Messages: []groqMessage{
			{
				Role: "user",
				Content: []groqContentItem{
					{Type: "text", Text: prompt},
					{
						Type: "image_url",
						ImageURL: &groqImageURL{
							URL: "data:image/jpeg;base64," + b64Data,
						},
					},
				},
			},
		},
		ResponseFormat: groqResponseFormat{
			Type: "json_object",
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("groq api call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return nil, fmt.Errorf("groq api returned status %d: %v", resp.StatusCode, errData)
	}

	var response groqResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("groq choices empty")
	}

	rawJSON := response.Choices[0].Message.Content

	var ktpData KTPData
	if err := json.Unmarshal([]byte(rawJSON), &ktpData); err != nil {
		return nil, fmt.Errorf("failed to parse extracted JSON: %w, raw response: %s", err, rawJSON)
	}

	return &ktpData, nil
}
