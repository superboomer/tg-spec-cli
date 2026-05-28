package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

// fakeBotAPIPage is a minimal but realistic Telegram Bot API documentation page:
// a version marker, a plain object type, a union type, and two methods (one with
// parameters, one without).
const fakeBotAPIPage = `<!DOCTYPE html><html><body>
	<strong>Bot API 7.0</strong>

	<h4>User</h4>
	<p>This object represents a Telegram user or bot.</p>
	<table>
		<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
		<tbody>
			<tr><td>id</td><td>Integer</td><td>Unique identifier.</td></tr>
			<tr><td>username</td><td>String</td><td>Optional. The username.</td></tr>
		</tbody>
	</table>

	<h4>ChatMember</h4>
	<p>This object contains information about one member of a chat. It can be one of</p>
	<ul>
		<li><a href="#chatmemberowner">ChatMemberOwner</a></li>
		<li><a href="#chatmembermember">ChatMemberMember</a></li>
	</ul>

	<h4>getMe</h4>
	<p>Returns basic information about the bot as a <a href="#user">User</a> object.</p>

	<h4>sendMessage</h4>
	<p>Use this method to send text messages.</p>
	<table>
		<thead><tr><th>Parameter</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
		<tbody>
			<tr><td>chat_id</td><td>Integer or String</td><td>Yes</td><td>Target chat.</td></tr>
			<tr><td>text</td><td>String</td><td>Yes</td><td>Message text.</td></tr>
		</tbody>
	</table>
</body></html>`

func newPageServer(t *testing.T, html string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(html))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestApp_Run_Success(t *testing.T) {
	srv := newPageServer(t, fakeBotAPIPage)
	out := filepath.Join(t.TempDir(), "spec-%v.json")

	a := NewWithType(zap.NewNop(), srv.URL, out, "botapi")
	if err := a.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	path := filepath.Join(filepath.Dir(out), "spec-7.0.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected generated spec at %s: %v", path, err)
	}
	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("generated spec is not valid JSON: %v", err)
	}
	paths, _ := spec["paths"].(map[string]any)
	if _, ok := paths["/getMe"]; !ok {
		t.Error("expected /getMe path in generated spec")
	}
	if _, ok := paths["/sendMessage"]; !ok {
		t.Error("expected /sendMessage path in generated spec")
	}
}

func TestApp_Run_DebugLogging(t *testing.T) {
	// Exercise the debug-level branches that enumerate type and method names.
	srv := newPageServer(t, fakeBotAPIPage)
	out := filepath.Join(t.TempDir(), "spec.json")

	log, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	a := NewWithType(log, srv.URL, out, "botapi")
	if err := a.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestApp_Run_UnreachableURL(t *testing.T) {
	a := NewWithType(zap.NewNop(), "http://127.0.0.1:0", "out.json", "botapi")
	if err := a.Run(); err == nil {
		t.Error("expected error for unreachable URL")
	}
}

func TestApp_Run_VersionError(t *testing.T) {
	// Page without any version markers -> GetVersion fails.
	srv := newPageServer(t, `<!DOCTYPE html><html><body><p>nothing</p></body></html>`)
	a := NewWithType(zap.NewNop(), srv.URL, "out.json", "botapi")
	if err := a.Run(); err == nil {
		t.Error("expected error when version cannot be determined")
	}
}

func TestApp_Run_SaveError(t *testing.T) {
	srv := newPageServer(t, fakeBotAPIPage)
	// Use an existing file as a directory component so Save's MkdirAll fails.
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	a := NewWithType(zap.NewNop(), srv.URL, filepath.Join(blocker, "sub", "out.json"), "botapi")
	if err := a.Run(); err == nil {
		t.Error("expected error when output cannot be saved")
	}
}
