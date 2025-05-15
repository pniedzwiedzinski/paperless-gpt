package ocr

import (
	"context"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "image/jpeg"
	
	"github.com/sirupsen/logrus"
)

// SuryaOCRProvider implements the ocr.Provider interface for the Surya OCR engine.
// It sends base64 encoded images to a configurable Surya endpoint for OCR processing.
// SuryaOCRProvider implements the ocr.Provider interface for the Surya OCR engine.
type SuryaOCRProvider struct {
	Endpoint string
	Token    string
}

// NewSuryaOCRProvider creates a new instance of SuryaOCRProvider.
// It reads the Surya endpoint and authentication token from environment variables.
func NewSuryaOCRProvider(cfg Config) (*SuryaOCRProvider, error) {
	log.Printf("Initializing Surya OCR provider")
	return &SuryaOCRProvider{
		Endpoint: cfg.SuryaEndpoint,
		Token:    cfg.SuryaToken,
	}
}

// ProcessImage processes the provided image data using the Surya OCR provider.
func (p *SuryaOCRProvider) ProcessImage(ctx context.Context, imageData []byte, page int) (*OCRResult, error) {
	log.Printf("Processing image with Surya OCR provider for page %d", page)

	if p.Endpoint == "" {
		return nil, fmt.Errorf("SURYA_ENDPOINT environment variable not set")
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	requestData := map[string]interface{}{
		"json": map[string]string{
			"mime_type": "image/jpeg", // Assuming image/jpeg for now
			"data":      base64Image,
		},
	}

	requestBody, err := json.Marshal(requestData)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.Endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.Token != "" {
		req.Header.Set("Authorization", "Bearer "+p.Token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("surya API returned non-200 status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read surya API response body: %w", err)
	}

	var suryaResponse struct {
		Text string `json:"text"`
	}

	err = json.Unmarshal(bodyBytes, &suryaResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal surya API response: %w, body: %s", err, string(bodyBytes))
	}

	ocrResult := &OCRResult{Text: suryaResponse.Text}

	log.Printf("Surya API response status code: %d", resp.StatusCode)
	return ocrResult, nil
}