package telegram

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// docFromHTML is a small helper for building a goquery document in tests.
func docFromHTML(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}
	return doc
}

func TestIsUnionDescription(t *testing.T) {
	tests := []struct {
		desc string
		want bool
	}{
		{"This object represents an incoming update.", true},
		{"This object describes the position.", true},
		{"This object contains information about a chat.", true},
		{"It can be one of the following.", true},
		{"Should be one of the listed types.", true},
		{"CASE INSENSITIVE: This Object Represents X", true},
		{"A regular object with fields.", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := IsUnionDescription(tt.desc); got != tt.want {
				t.Errorf("IsUnionDescription(%q) = %v, want %v", tt.desc, got, tt.want)
			}
		})
	}
}

func TestIsMethodName(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"getMe", true},
		{"sendMessage", true},
		{"Message", false},        // type (uppercase)
		{"Recent Changes", false}, // section (space)
		{"", false},
		{"with\ttab", false},
		{"123abc", false}, // not a letter first
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := isMethodName(tt.in); got != tt.want {
				t.Errorf("isMethodName(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsFirstLetterUppercase(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"User", true},
		{"message", false},
		{"", false},
		{"123", false},
		{" Leading", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := isFirstLetterUppercase(tt.in); got != tt.want {
				t.Errorf("isFirstLetterUppercase(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseDataType_FromLinks(t *testing.T) {
	doc := docFromHTML(t, `<table><tr><td>Array of <a href="#user">User</a></td></tr></table>`)
	page := &PageAPI{Types: map[string]Type{"user": {Name: "User"}}}
	dt := page.parseDataType(doc.Find("td"))
	if !dt.IsArray || dt.ArrayDepth != 1 {
		t.Errorf("expected array depth 1, got IsArray=%v depth=%d", dt.IsArray, dt.ArrayDepth)
	}
	if len(dt.Types) != 1 || dt.Types[0] != "User" {
		t.Errorf("expected [User] resolved from link, got %v", dt.Types)
	}
}

func TestParseDataType_MixedRefAndPrimitive(t *testing.T) {
	// "InputFile or String": the linked type and the bare primitive must both
	// be captured (regression: the primitive used to be dropped).
	doc := docFromHTML(t, `<table><tr><td><a href="#inputfile">InputFile</a> or String</td></tr></table>`)
	page := &PageAPI{Types: map[string]Type{"inputfile": {Name: "InputFile"}}}
	dt := page.parseDataType(doc.Find("td"))
	if len(dt.Types) != 2 || dt.Types[0] != "InputFile" || dt.Types[1] != "String" {
		t.Errorf("expected [InputFile String], got %v", dt.Types)
	}
}

func TestParseDataType_NestedArray(t *testing.T) {
	doc := docFromHTML(t, `<table><tr><td>Array of Array of String</td></tr></table>`)
	page := &PageAPI{}
	dt := page.parseDataType(doc.Find("td"))
	if dt.ArrayDepth != 2 {
		t.Errorf("expected ArrayDepth 2, got %d", dt.ArrayDepth)
	}
	if len(dt.Types) != 1 || dt.Types[0] != "String" {
		t.Errorf("expected [String], got %v", dt.Types)
	}
}

func TestParseDataType_LinkToUnknownTypeIsSkipped(t *testing.T) {
	// A link whose type isn't loaded is ignored, falling back to text parsing.
	doc := docFromHTML(t, `<table><tr><td><a href="#missing">Missing</a></td></tr></table>`)
	page := &PageAPI{Types: map[string]Type{}}
	dt := page.parseDataType(doc.Find("td"))
	if len(dt.Types) != 1 || dt.Types[0] != "Missing" {
		t.Errorf("expected fallback to text [Missing], got %v", dt.Types)
	}
}

func TestGetType_EmptyMap(t *testing.T) {
	page := &PageAPI{Types: map[string]Type{}}
	if _, err := page.GetType("anything"); err == nil {
		t.Error("expected error when types map is empty")
	}
}

func TestGetTypes_LazyLoads(t *testing.T) {
	doc := docFromHTML(t, `<!DOCTYPE html><html><body>
		<h4>User</h4>
		<table>
			<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
			<tbody><tr><td>id</td><td>Integer</td><td>The id</td></tr></tbody>
		</table>
	</body></html>`)
	page := &PageAPI{Document: doc, Types: make(map[string]Type)}
	types, err := page.GetTypes()
	if err != nil {
		t.Fatalf("GetTypes() error = %v", err)
	}
	if _, ok := types["user"]; !ok {
		t.Errorf("expected lazily-loaded 'user' type, got %v", types)
	}
}

func TestLoadTypes_RequiredFromOptionalPrefix(t *testing.T) {
	doc := docFromHTML(t, `<!DOCTYPE html><html><body>
		<h4>User</h4>
		<table>
			<thead><tr><th>Field</th><th>Type</th><th>Description</th></tr></thead>
			<tbody>
				<tr><td>id</td><td>Integer</td><td>Unique identifier.</td></tr>
				<tr><td>username</td><td>String</td><td>Optional. The username.</td></tr>
			</tbody>
		</table>
	</body></html>`)
	page := &PageAPI{Document: doc, Types: make(map[string]Type)}
	if err := page.LoadTypes(); err != nil {
		t.Fatal(err)
	}
	user := page.Types["user"]
	got := map[string]bool{}
	for _, f := range user.Fields {
		got[f.Name] = f.Required
	}
	if !got["id"] {
		t.Error("id should be required (no 'Optional.' prefix)")
	}
	if got["username"] {
		t.Error("username should be optional (has 'Optional.' prefix)")
	}
}

func TestLoadTypes_RequiredColumn(t *testing.T) {
	doc := docFromHTML(t, `<!DOCTYPE html><html><body>
		<h4>Sample</h4>
		<table>
			<thead><tr><th>Field</th><th>Type</th><th>Required</th><th>Description</th></tr></thead>
			<tbody>
				<tr><td>a</td><td>String</td><td>Yes</td><td>desc</td></tr>
				<tr><td>b</td><td>String</td><td>No</td><td>desc</td></tr>
			</tbody>
		</table>
	</body></html>`)
	page := &PageAPI{Document: doc, Types: make(map[string]Type)}
	if err := page.LoadTypes(); err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, f := range page.Types["sample"].Fields {
		got[f.Name] = f.Required
	}
	if !got["a"] || got["b"] {
		t.Errorf("expected a=required, b=optional from Required column, got %v", got)
	}
}

func TestLoadTypes_UnionAndPlaceholder(t *testing.T) {
	doc := docFromHTML(t, `<!DOCTYPE html><html><body>
		<h4>ChatMember</h4>
		<p>This object contains information about a chat member. It can be one of</p>
		<ul>
			<li><a href="#chatmemberowner">ChatMemberOwner</a></li>
			<li><a href="#chatmembermember">ChatMemberMember</a></li>
		</ul>
		<h4>CallbackGame</h4>
		<p>A placeholder, currently holds no information.</p>
		<h4>Ignored Section</h4>
		<p>Just prose, not a type.</p>
	</body></html>`)
	page := &PageAPI{Document: doc, Types: make(map[string]Type)}
	if err := page.LoadTypes(); err != nil {
		t.Fatal(err)
	}

	cm, ok := page.Types["chatmember"]
	if !ok {
		t.Fatal("union type ChatMember should be kept")
	}
	if len(cm.Fields) != 0 {
		t.Errorf("union type should have no fields, got %v", cm.Fields)
	}
	if !strings.Contains(cm.Description, "- ChatMemberOwner") {
		t.Errorf("union variants should be appended to description: %q", cm.Description)
	}
	if _, ok := page.Types["callbackgame"]; !ok {
		t.Error("placeholder type CallbackGame should be kept")
	}
	if _, ok := page.Types["ignored section"]; ok {
		t.Error("prose-only heading should not be kept as a type")
	}
}

func TestGetMethods_ReturnTypes(t *testing.T) {
	doc := docFromHTML(t, `<!DOCTYPE html><html><body>
		<h4>getChat</h4>
		<p>Returns a <a href="#chat">Chat</a> object on success.</p>
		<h4>getUpdates</h4>
		<p>Returns an Array of <a href="#update">Update</a> objects.</p>
		<h4>deleteMessage</h4>
		<p>Returns <em>True</em> on success.</p>
		<h4>getCount</h4>
		<p>Returns the new <em>Int</em> value.</p>
	</body></html>`)
	page := &PageAPI{
		Document: doc,
		Types: map[string]Type{
			"chat":   {Name: "Chat"},
			"update": {Name: "Update"},
		},
	}
	methods, err := page.GetMethods()
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Method{}
	for _, m := range methods {
		byName[m.Name] = m
	}

	if rt := byName["getChat"].ReturnType; rt.Name != "Chat" || rt.IsArray {
		t.Errorf("getChat return = %+v, want {Chat false}", rt)
	}
	if rt := byName["getUpdates"].ReturnType; rt.Name != "Update" || !rt.IsArray {
		t.Errorf("getUpdates return = %+v, want {Update true}", rt)
	}
	if rt := byName["deleteMessage"].ReturnType; rt.Name != "boolean" {
		t.Errorf("deleteMessage return = %+v, want boolean", rt)
	}
	if rt := byName["getCount"].ReturnType; rt.Name != "integer" {
		t.Errorf("getCount return = %+v, want integer", rt)
	}
}

func TestGetVersion_PicksFirstStrong(t *testing.T) {
	doc := docFromHTML(t, `<!DOCTYPE html><html><body>
		<strong>Bot API 7.0</strong>
		<strong>Bot API 6.9</strong>
	</body></html>`)
	page := &PageAPI{Document: doc}
	v, err := page.GetVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != "7.0" {
		t.Errorf("expected first version 7.0, got %q", v)
	}
}

func TestGetVersion_RecentChangesWithoutH4(t *testing.T) {
	// "Recent Changes" present but no following h4 -> specific error.
	doc := docFromHTML(t, `<!DOCTYPE html><html><body>
		<h3>Recent Changes</h3>
		<p>No version heading here.</p>
	</body></html>`)
	page := &PageAPI{Document: doc}
	if _, err := page.GetVersion(); err == nil {
		t.Error("expected error when no h4 follows Recent Changes")
	}
}
