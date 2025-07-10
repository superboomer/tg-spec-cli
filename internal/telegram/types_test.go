package telegram

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestPageAPI_GetType(t *testing.T) {
	html := `<!DOCTYPE html>
	<html>
		<body>
			<h4>TestType</h4>
			<table>
				<tr>
					<td>field1</td>
					<td>String</td>
					<td>Test description</td>
				</tr>
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
			"TestType": {
				Name: "TestType",
				Fields: []Field{
					{
						Name:        "field1",
						Type:        []string{"String"},
						Description: "Test description",
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		typeName string
		want     Type
		wantErr  bool
	}{
		{
			name:     "existing type",
			typeName: "TestType",
			want: Type{
				Name: "TestType",
				Fields: []Field{
					{
						Name:        "field1",
						Type:        []string{"String"},
						Description: "Test description",
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "non-existing type",
			typeName: "NonExistingType",
			want:     Type{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pageAPI.GetType(tt.typeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("PageAPI.GetType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name != tt.want.Name {
				t.Errorf("PageAPI.GetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageAPI_LoadTypes(t *testing.T) {
	html := `<!DOCTYPE html>
	<html>
		<body>
			<h4>Type1</h4>
			<table>
				<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
				<tbody><tr><td>field1</td><td>String</td><td>Description 1</td></tr></tbody>
			</table>
			<h4>Type2</h4>
			<table>
				<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
				<tbody><tr><td>field2</td><td>Integer</td><td>Description 2</td></tr></tbody>
			</table>
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}

	pageAPI := &PageAPI{
		Document: doc,
		Types:    make(map[string]Type),
	}

	if err := pageAPI.LoadTypes(); err != nil {
		t.Errorf("PageAPI.LoadTypes() error = %v", err)
		return
	}

	// Verify that types were loaded
	if len(pageAPI.Types) != 2 {
		t.Errorf("Expected 2 types loaded, got %d", len(pageAPI.Types))
	}
}
