package telegram

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestPageAPI_GetMethods(t *testing.T) {
	html := `<!DOCTYPE html>
	<html>
		<body>
			<h4>TestMethod</h4>
			<p>Returns <a href=\"#TestType\">TestType</a></p>
			<table>
				<thead><tr><th>Parameter</th><th>Type</th><th>Description</th><th>Required</th></tr></thead>
				<tbody><tr><td>param1</td><td>String</td><td>Description</td><td>Yes</td></tr></tbody>
			</table>
			<h4>Dummy</h4>
		</body>
	</html>`
	fmt.Println("DEBUG: HTML input:\n", html)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}
	fmt.Println("DEBUG: doc.Find('h4, table').Length():", doc.Find("h4, table").Length())
	pageAPI := &PageAPI{
		Document: doc,
		Types: map[string]Type{
			"TestType": {Name: "TestType"},
		},
	}
	methods, err := pageAPI.GetMethods()
	fmt.Println("DEBUG: methods:", methods)
	if err != nil {
		t.Errorf("GetMethods() error = %v", err)
	}
	if len(methods) == 0 {
		t.Fatalf("GetMethods() returned no methods")
	}
	if methods[0].Name != "TestMethod" {
		t.Errorf("Expected method name 'TestMethod', got %v", methods[0].Name)
	}
	if len(methods[0].Parameters) == 0 {
		t.Errorf("Expected at least one parameter, got %v", methods[0].Parameters)
	}
}
