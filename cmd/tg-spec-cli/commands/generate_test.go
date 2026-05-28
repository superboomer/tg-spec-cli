package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// fakeBotAPIPage is a minimal documentation page sufficient to drive a full
// generate run (version marker + one type + one method).
const fakeBotAPIPage = `<!DOCTYPE html><html><body>
	<strong>Bot API 7.0</strong>
	<h4>User</h4>
	<p>This object represents a user.</p>
	<table>
		<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
		<tbody><tr><td>id</td><td>Integer</td><td>Unique identifier.</td></tr></tbody>
	</table>
	<h4>getMe</h4>
	<p>Returns basic information about the bot as a <a href="#user">User</a> object.</p>
</body></html>`

func pageServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(fakeBotAPIPage))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestGenerateCmd(t *testing.T) {
	cmd := generateCmd
	if cmd.Use != "generate" {
		t.Errorf("generateCmd.Use = %v, want 'generate'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("generateCmd.Short should not be empty")
	}
}

func TestGenerateCmdRun_BotAPI(t *testing.T) {
	srv := pageServer(t)
	dir := t.TempDir()
	outputPath = filepath.Join(dir, "spec-%v.json")
	logLevel = "info"
	url = srv.URL
	typeFlag = "botapi"

	generateCmd.Run(&cobra.Command{}, []string{})

	if _, err := os.Stat(filepath.Join(dir, "spec-7.0.json")); err != nil {
		t.Errorf("expected generated spec file: %v", err)
	}
}

func TestGenerateCmdRun_GatewayKeepsExplicitURL(t *testing.T) {
	srv := pageServer(t)
	dir := t.TempDir()
	outputPath = filepath.Join(dir, "gw.json")
	logLevel = "debug"
	url = srv.URL
	typeFlag = "gateway"

	// Provide a command whose "url" flag is explicitly changed so the gateway
	// branch does NOT overwrite our local test URL.
	cmd := &cobra.Command{}
	cmd.Flags().String("url", "", "")
	if err := cmd.Flags().Set("url", srv.URL); err != nil {
		t.Fatal(err)
	}

	generateCmd.Run(cmd, []string{})

	if _, err := os.Stat(filepath.Join(dir, "gw.json")); err != nil {
		t.Errorf("expected generated gateway spec file: %v", err)
	}
}

func TestGenerateCmdRun_InvalidLogLevel(t *testing.T) {
	// Invalid log level must return early (and must not panic on a nil logger).
	logLevel = "definitely-not-a-level"
	url = "http://127.0.0.1:0"
	typeFlag = "botapi"
	outputPath = filepath.Join(t.TempDir(), "out.json")

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("generateCmd.Run panicked on invalid log level: %v", r)
		}
	}()
	generateCmd.Run(&cobra.Command{}, []string{})
}
