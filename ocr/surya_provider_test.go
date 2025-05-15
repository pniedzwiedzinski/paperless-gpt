package ocr_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/paperless-gpt/paperless-gpt/pkg/ocr"
)

func TestSuryaOCRProvider_ProcessImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic assertions about the request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test_token" {
			t.Errorf("Expected Authorization header 'Bearer test_token', got %s", r.Header.Get("Authorization"))
		}
		// We can add more checks for the request body later if needed

		// Respond with a dummy success response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{
			"text": "This is a test document.",
			"lines": [
				{"text": "This is a", "box": [10, 10, 50, 20]},
				{"text": "test document.", "box": [60, 10, 100, 20]},
				{"text": "Another line.", "box": [10, 30, 60, 40]}
			]
		}`)
	}))
	defer server.Close()

	// Dummy image data
	imageData := []byte("dummy image data")
	page := 1

	// Set environment variables for the provider
	os.Setenv("SURYA_ENDPOINT", server.URL)
	os.Setenv("SURYA_AUTH_TOKEN", "test_token")
	defer os.Unsetenv("SURYA_ENDPOINT")
	defer os.Unsetenv("SURYA_AUTH_TOKEN")

	cfg := ocr.Config{} // Dummy config for now
	provider := ocr.NewSuryaOCRProvider(cfg)

	result, err = provider.ProcessImage(context.Background(), imageData, page)

	if err != nil {
		t.Errorf("ProcessImage returned an error: %v", err)
	}

	expectedText := "This is a test document."
	if result.Text != expectedText {
		t.Errorf("ProcessImage returned unexpected text. Got: %q, Expected: %q", result.Text, expectedText)
	}

	expectedLines := 3
	if len(result.Lines) != expectedLines {
		t.Errorf("ProcessImage returned unexpected number of lines. Got: %d, Expected: %d", len(result.Lines), expectedLines)
	}
}