package telegram

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseDataType(t *testing.T) {
	// Test with array and multiple types (plain text)
	html := `<table><tr><td>Array of String or Integer</td></tr></table>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}
	pageAPI := &PageAPI{}
	cell := doc.Find("td")
	fmt.Println("DEBUG: doc.Text() for array:", cell.Text())
	dt := pageAPI.parseDataType(cell)
	if !dt.IsArray {
		t.Errorf("Expected IsArray to be true, got false")
	}
	if dt.ArrayDepth != 1 {
		t.Errorf("Expected ArrayDepth 1, got %d", dt.ArrayDepth)
	}
	if len(dt.Types) != 2 || dt.Types[0] != "String" || dt.Types[1] != "Integer" {
		t.Errorf("Expected Types [String Integer], got %v", dt.Types)
	}

	// Test with no array, multiple types
	html2 := `<table><tr><td>String or Integer</td></tr></table>`
	doc2, _ := goquery.NewDocumentFromReader(strings.NewReader(html2))
	cell2 := doc2.Find("td")
	fmt.Println("DEBUG: doc.Text() for multi:", cell2.Text())
	dt2 := pageAPI.parseDataType(cell2)
	if len(dt2.Types) != 2 || dt2.Types[0] != "String" || dt2.Types[1] != "Integer" {
		t.Errorf("Expected Types [String Integer], got %v", dt2.Types)
	}

	// Test with a single type, no array
	html3 := `<table><tr><td>Boolean</td></tr></table>`
	doc3, _ := goquery.NewDocumentFromReader(strings.NewReader(html3))
	cell3 := doc3.Find("td")
	fmt.Println("DEBUG: doc.Text() for single:", cell3.Text())
	dt3 := pageAPI.parseDataType(cell3)
	if len(dt3.Types) != 1 || dt3.Types[0] != "Boolean" {
		t.Errorf("Expected Types [Boolean], got %v", dt3.Types)
	}
}
