package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/superboomer/tg-spec-cli/internal/openapi"
	"github.com/superboomer/tg-spec-cli/internal/telegram"
	"go.uber.org/zap"
)

// sampleTypes returns a representative set of parsed types: a plain object with
// required and optional fields, and a union type.
func sampleTypes() map[string]telegram.Type {
	return map[string]telegram.Type{
		"user": {
			Name:        "User",
			Description: "This object represents a user.",
			Fields: []telegram.Field{
				{Name: "id", Type: []string{"Integer"}, Description: "Unique id", Required: true},
				{Name: "username", Type: []string{"String"}, Description: "Optional. Username", Required: false},
			},
		},
		"chatmember": {
			Name:        "ChatMember",
			Description: "This object contains info about a chat member. It can be one of\n- ChatMemberOwner\n- ChatMemberMember\n",
		},
	}
}

func sampleMethods() []telegram.Method {
	return []telegram.Method{
		{
			Name:        "getMe",
			Description: "Returns basic information about the bot.",
			ReturnType:  telegram.ReturnType{Name: "User"},
		},
		{
			Name:        "sendMessage",
			Description: "Use this method to send text messages.",
			ReturnType:  telegram.ReturnType{Name: "Message"},
			Parameters: []telegram.Parameter{
				{Name: "chat_id", Type: telegram.DataType{Types: []string{"Integer", "String"}}, Required: true},
				{Name: "text", Type: telegram.DataType{Types: []string{"String"}}, Required: true},
			},
		},
	}
}

func TestGenerate_BotAPI(t *testing.T) {
	gen := NewWithType(zap.NewNop(), "7.0", sampleTypes(), sampleMethods(), "botapi")
	spec, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if spec.OpenAPI != "3.1.0" {
		t.Errorf("OpenAPI version = %q, want 3.1.0", spec.OpenAPI)
	}
	if spec.Info.Title != "Telegram Bot API" {
		t.Errorf("Info.Title = %q", spec.Info.Title)
	}
	if spec.Info.Version != "7.0" {
		t.Errorf("Info.Version = %q, want 7.0", spec.Info.Version)
	}
	if len(spec.Servers) != 2 {
		t.Errorf("expected 2 servers for botapi, got %d", len(spec.Servers))
	}
	if spec.Servers[0].Variables == nil || spec.Servers[0].Variables.Token == nil {
		t.Error("botapi server should have a token variable")
	}
	if spec.Security != nil {
		t.Errorf("botapi must not declare global security, got %v", spec.Security)
	}
	if spec.Components.SecuritySchemes != nil {
		t.Errorf("botapi must not declare security schemes, got %v", spec.Components.SecuritySchemes)
	}

	// Plain object: required list reflects field.Required.
	user, ok := spec.Components.Schemas["User"]
	if !ok {
		t.Fatal("User schema missing")
	}
	if len(user.Required) != 1 || user.Required[0] != "id" {
		t.Errorf("User.Required = %v, want [id]", user.Required)
	}
	if _, ok := user.Properties["username"]; !ok {
		t.Error("User.Properties should contain username")
	}

	// Union type: oneOf populated, and no bogus type:object.
	cm, ok := spec.Components.Schemas["ChatMember"]
	if !ok {
		t.Fatal("ChatMember schema missing")
	}
	if cm.Type != "" {
		t.Errorf("union schema must not set type, got %q", cm.Type)
	}
	if len(cm.OneOf) != 2 {
		t.Errorf("ChatMember.OneOf = %v, want 2 variants", cm.OneOf)
	}

	// Paths: one POST per method, response includes ok + result.
	op, ok := spec.Paths["/sendMessage"]
	if !ok {
		t.Fatal("/sendMessage path missing")
	}
	props := op.Post.RequestBody.Content.Applicationjson.Schema.Properties
	if _, ok := props["chat_id"]; !ok {
		t.Error("sendMessage should expose chat_id parameter")
	}
	resp200 := op.Post.Responses["200"].Content.Applicationjson.Schema
	if _, ok := resp200.Properties["ok"]; !ok {
		t.Error("response schema must include 'ok'")
	}
	if _, ok := spec.Paths["/getMe"]; !ok {
		t.Error("parameterless method getMe should still produce a path")
	}
}

func TestGenerate_Gateway(t *testing.T) {
	gen := NewWithType(zap.NewNop(), "2025", sampleTypes(), sampleMethods(), "gateway")
	spec, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if spec.Info.Title != "Telegram Gateway API" {
		t.Errorf("Info.Title = %q", spec.Info.Title)
	}
	if len(spec.Servers) != 1 || spec.Servers[0].URL != "https://gatewayapi.telegram.org/" {
		t.Errorf("unexpected gateway servers: %+v", spec.Servers)
	}
	if spec.Servers[0].Variables != nil {
		t.Errorf("gateway server must not emit variables, got %+v", spec.Servers[0].Variables)
	}
	if spec.Components.SecuritySchemes["access_token"].Scheme != "bearer" {
		t.Error("gateway must define a bearer access_token security scheme")
	}
	if len(spec.Security) != 1 {
		t.Errorf("gateway must declare a global security requirement, got %v", spec.Security)
	}
}

func TestGenerate_DefaultTypeFlag(t *testing.T) {
	// Empty type flag falls through to the botapi branch.
	gen := NewWithType(zap.NewNop(), "1.0", map[string]telegram.Type{}, nil, "")
	spec, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if spec.Info.Title != "Telegram Bot API" {
		t.Errorf("empty type should default to Bot API, got %q", spec.Info.Title)
	}
}

func TestGenerate_UnknownType(t *testing.T) {
	gen := NewWithType(zap.NewNop(), "1.0", map[string]telegram.Type{}, nil, "weird")
	if _, err := gen.Generate(); err == nil {
		t.Error("Generate() with unknown type should return an error")
	}
}

func TestSave(t *testing.T) {
	gen := NewWithType(zap.NewNop(), "7.0", nil, nil, "botapi")
	spec := &openapi.OpenAPI{OpenAPI: "3.1.0"}

	t.Run("directory uses default name with version", func(t *testing.T) {
		dir := t.TempDir()
		if err := gen.Save(spec, dir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		assertValidSpec(t, filepath.Join(dir, "openapi-v7.0.json"))
	})

	t.Run("trailing slash treated as directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := gen.Save(spec, dir+"/"); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		assertValidSpec(t, filepath.Join(dir, "openapi-v7.0.json"))
	})

	t.Run("explicit file with version placeholder", func(t *testing.T) {
		dir := t.TempDir()
		if err := gen.Save(spec, filepath.Join(dir, "bot-api-%v.json")); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		assertValidSpec(t, filepath.Join(dir, "bot-api-7.0.json"))
	})

	t.Run("explicit file without placeholder", func(t *testing.T) {
		dir := t.TempDir()
		if err := gen.Save(spec, filepath.Join(dir, "exact.json")); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		assertValidSpec(t, filepath.Join(dir, "exact.json"))
	})

	t.Run("creates nested directories", func(t *testing.T) {
		dir := t.TempDir()
		if err := gen.Save(spec, filepath.Join(dir, "a", "b", "spec-%v.json")); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		assertValidSpec(t, filepath.Join(dir, "a", "b", "spec-7.0.json"))
	})

	t.Run("version with comma and spaces is sanitized", func(t *testing.T) {
		dir := t.TempDir()
		g := NewWithType(zap.NewNop(), "February 26, 2025", nil, nil, "gateway")
		if err := g.Save(spec, filepath.Join(dir, "gw-%v.json")); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		assertValidSpec(t, filepath.Join(dir, "gw-February26-2025.json"))
	})

	t.Run("returns error when directory cannot be created", func(t *testing.T) {
		dir := t.TempDir()
		// Create a regular file, then try to nest a path beneath it.
		blocker := filepath.Join(dir, "file")
		if err := os.WriteFile(blocker, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
		err := gen.Save(spec, filepath.Join(blocker, "sub", "out.json"))
		if err == nil {
			t.Error("Save() should fail when output directory cannot be created")
		}
	})
}

func assertValidSpec(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected file at %s: %v", path, err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Errorf("file %s is not valid JSON: %v", path, err)
	}
}
