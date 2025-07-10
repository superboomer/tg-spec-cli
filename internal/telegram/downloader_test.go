package telegram

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPage(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantErr bool
	}{
		{
			name: "valid html",
			html: `<!DOCTYPE html>
			<html>
				<body>
					<h4>Test Type</h4>
					<table>
						<tr><td>Field</td><td>Type</td><td>Description</td></tr>
					</table>
				</body>
			</html>`,
			wantErr: false,
		},
		{
			name:    "empty html",
			html:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.html))
			}))
			defer server.Close()

			got, err := GetPage(server.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("GetPage() returned nil but no error")
			}
		})
	}

	// Test error cases
	t.Run("invalid url", func(t *testing.T) {
		_, err := GetPage("invalid-url")
		if err == nil {
			t.Error("GetPage() with invalid URL should return error")
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		_, err := GetPage(server.URL)
		if err == nil {
			t.Error("GetPage() with server error should return error")
		}
	})
}
