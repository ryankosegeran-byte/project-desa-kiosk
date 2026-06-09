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

type GeminiProvider struct {
	APIKey    string
	ModelName string // default: gemini-1.5-flash
}

func NewGeminiProvider(apiKey, modelName string) *GeminiProvider {
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}
	return &GeminiProvider{APIKey: apiKey, ModelName: modelName}
}

func (p *GeminiProvider) Name() string { return "gemini" }

// API Request/Response Types
type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string `json:"responseMimeType"`
}

type geminiRequestPayload struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiResponsePayload struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (p *GeminiProvider) ExtractKTP(ctx context.Context, imageData []byte) (*KTPData, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("API Key Gemini belum diatur")
	}

	b64Data := base64.StdEncoding.EncodeToString(imageData)

	prompt := `Ekstrak informasi dari foto KTP-el Indonesia ini ke dalam format JSON yang valid.
Kembalikan HANYA objek JSON tanpa format markdown (no backticks).
Skema JSON:
{
  "nik": "string NIK 16 digit",
  "nama": "string nama lengkap",
  "tempat_lahir": "string tempat lahir",
  "tanggal_lahir": "string format YYYY-MM-DD",
  "jenis_kelamin": "string L atau P",
  "alamat": "string alamat lengkap tanpa RT/RW",
  "rt": "string RT saja",
  "rw": "string RW saja",
  "kelurahan": "string kelurahan/desa",
  "kecamatan": "string kecamatan",
  "agama": "string agama",
  "status_kawin": "string status perkawinan",
  "pekerjaan": "string pekerjaan",
  "kewarganegaraan": "string WNI atau WNA",
  "confidence": "float nilai keyakinan 0.0 - 1.0"
}`

	payload := geminiRequestPayload{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
					{
						InlineData: &geminiInlineData{
							MimeType: "image/jpeg",
							Data:     b64Data,
						},
					},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.ModelName, p.APIKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini api call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return nil, fmt.Errorf("gemini api returned status %d: %v", resp.StatusCode, errData)
	}

	var response geminiResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini response candidates empty")
	}

	rawJSON := response.Candidates[0].Content.Parts[0].Text

	var ktpData KTPData
	if err := json.Unmarshal([]byte(rawJSON), &ktpData); err != nil {
		return nil, fmt.Errorf("failed to parse extracted JSON: %w, raw response: %s", err, rawJSON)
	}

	return &ktpData, nil
}
