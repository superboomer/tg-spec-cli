package telegram

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestPageAPI_GetMethods(t *testing.T) {
	html := `<!DOCTYPE html>
	<html>
		<body>
			<h4>testMethod</h4>
			<p>Returns <a href="#TestType">TestType</a></p>
			<table>
				<thead><tr><th>Parameter</th><th>Type</th><th>Description</th><th>Required</th></tr></thead>
				<tbody><tr><td>param1</td><td>String</td><td>Description</td><td>Yes</td></tr></tbody>
			</table>
			<h4>getMe</h4>
			<p>A simple method for testing.</p>
			<h4>SomeType</h4>
			<table>
				<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
				<tbody><tr><td>field1</td><td>String</td><td>Description</td></tr></tbody>
			</table>
		</body>
	</html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}
	pageAPI := &PageAPI{
		Document: doc,
		Types: map[string]Type{
			"TestType": {Name: "TestType"},
		},
	}
	methods, err := pageAPI.GetMethods()
	if err != nil {
		t.Errorf("GetMethods() error = %v", err)
	}
	if len(methods) != 2 {
		t.Fatalf("Expected 2 methods (including the parameterless one), got %d: %v", len(methods), methods)
	}
	if methods[0].Name != "testMethod" {
		t.Errorf("Expected method name 'testMethod', got %v", methods[0].Name)
	}
	if len(methods[0].Parameters) == 0 {
		t.Errorf("Expected at least one parameter, got %v", methods[0].Parameters)
	}
	// Parameterless methods (e.g. getMe) must still be captured.
	if methods[1].Name != "getMe" {
		t.Errorf("Expected parameterless method 'getMe', got %v", methods[1].Name)
	}
	// Type headings (uppercase) must not be treated as methods.
	for _, m := range methods {
		if m.Name == "SomeType" {
			t.Errorf("Type heading 'SomeType' was incorrectly parsed as a method")
		}
	}
}
