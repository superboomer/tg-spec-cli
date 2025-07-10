package telegram

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestPageAPI_GetVersion(t *testing.T) {
	html := `<!DOCTYPE html>
	<html>
		<body>
			<strong>Bot API 7.0</strong>
		</body>
	</html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}
	pageAPI := &PageAPI{Document: doc}
	version, err := pageAPI.GetVersion()
	if err != nil {
		t.Errorf("GetVersion() error = %v", err)
	}
	if version != "7.0" {
		t.Errorf("Expected version '7.0', got %v", version)
	}

	// Test fallback to Recent Changes
	html2 := `<!DOCTYPE html>
	<html>
		<body>
			<h3>Recent Changes</h3>
			<h4>6.9</h4>
		</body>
	</html>`
	doc2, _ := goquery.NewDocumentFromReader(strings.NewReader(html2))
	pageAPI2 := &PageAPI{Document: doc2}
	version2, err2 := pageAPI2.GetVersion()
	if err2 != nil {
		t.Errorf("GetVersion() error = %v", err2)
	}
	if version2 != "6.9" {
		t.Errorf("Expected version '6.9', got %v", version2)
	}

	// Test error case
	html3 := `<!DOCTYPE html><html><body></body></html>`
	doc3, _ := goquery.NewDocumentFromReader(strings.NewReader(html3))
	pageAPI3 := &PageAPI{Document: doc3}
	_, err3 := pageAPI3.GetVersion()
	if err3 == nil {
		t.Error("Expected error for missing version")
	}
}
