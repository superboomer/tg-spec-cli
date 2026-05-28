package telegram

import "testing"

func TestGetPage_URLValidation(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"unsupported scheme", "ftp://example.com"},
		{"missing host", "http://"},
		{"unparsable url", "http://[::1]:namedport"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := GetPage(tt.url); err == nil {
				t.Errorf("GetPage(%q) expected an error", tt.url)
			}
		})
	}
}

func TestGetTypes_AlreadyLoaded(t *testing.T) {
	// When types are already populated, GetTypes returns them without parsing
	// (Document is nil here, so any parsing attempt would panic).
	page := &PageAPI{Types: map[string]Type{"user": {Name: "User"}}}
	types, err := page.GetTypes()
	if err != nil {
		t.Fatalf("GetTypes() error = %v", err)
	}
	if len(types) != 1 {
		t.Errorf("expected 1 preloaded type, got %d", len(types))
	}
}
