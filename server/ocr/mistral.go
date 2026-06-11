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

type MistralProvider struct {
	APIKey    string
	ModelName string // default: pixtral-12b
}

func NewMistralProvider(apiKey, modelName string) *MistralProvider {
	if modelName == "" {
		modelName = "pixtral-12b"
	}
	return &MistralProvider{APIKey: apiKey, ModelName: modelName}
}

func (p *MistralProvider) Name() string       { return "mistral" }
func (p *MistralProvider) Model() string      { return p.ModelName }
func (p *MistralProvider) IsConfigured() bool { return p.APIKey != "" }

// API Request/Response Types
type mistralImageURL struct {
	URL string `json:"url"`
}

type mistralContentItem struct {
	Type     string           `json:"type"`
	Text     string           `json:"text,omitempty"`
	ImageURL *mistralImageURL `json:"image_url,omitempty"`
}

type mistralMessage struct {
	Role    string               `json:"role"`
	Content []mistralContentItem `json:"content"`
}

type mistralResponseFormat struct {
	Type string `json:"type"`
}

type mistralRequestPayload struct {
	Model          string                `json:"model"`
	Messages       []mistralMessage      `json:"messages"`
	ResponseFormat mistralResponseFormat `json:"response_format"`
}

type mistralResponsePayload struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *MistralProvider) ExtractKTP(ctx context.Context, imageData []byte) (*KTPData, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("API Key Mistral belum diatur")
	}

	b64Data := base64.StdEncoding.EncodeToString(imageData)

	prompt := `Ekstrak data dari gambar KTP-el berikut ke format JSON.
Kembalikan objek JSON dengan field: nik, nama, tempat_lahir, tanggal_lahir (YYYY-MM-DD), jenis_kelamin (L/P), alamat, rt, rw, kelurahan, kecamatan, agama, status_kawin, pekerjaan, kewarganegaraan, confidence.`

	payload := mistralRequestPayload{
		Model: p.ModelName,
		Messages: []mistralMessage{
			{
				Role: "user",
				Content: []mistralContentItem{
					{Type: "text", Text: prompt},
					{
						Type: "image_url",
						ImageURL: &mistralImageURL{
							URL: "data:image/jpeg;base64," + b64Data,
						},
					},
				},
			},
		},
		ResponseFormat: mistralResponseFormat{
			Type: "json_object",
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mistral api call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return nil, fmt.Errorf("mistral api returned status %d: %v", resp.StatusCode, errData)
	}

	var response mistralResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("mistral choices empty")
	}

	rawJSON := response.Choices[0].Message.Content

	var ktpData KTPData
	if err := json.Unmarshal([]byte(rawJSON), &ktpData); err != nil {
		return nil, fmt.Errorf("failed to parse extracted JSON: %w, raw response: %s", err, rawJSON)
	}

	return &ktpData, nil
}
