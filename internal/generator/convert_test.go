package generator

import (
	"reflect"
	"testing"

	"github.com/superboomer/tg-spec-cli/internal/telegram"
	"go.uber.org/zap"
)

// newTestGen returns a botapi generator wired with a no-op logger, used to
// exercise the type-conversion helpers.
func newTestGen() *Generator {
	return NewWithType(zap.NewNop(), "1.0", map[string]telegram.Type{}, []telegram.Method{}, "botapi")
}

func TestConvertType(t *testing.T) {
	g := newTestGen()
	tests := []struct {
		in   string
		want string
	}{
		{"Integer", "integer"},
		{"Int", "integer"},
		{"Float", "number"},
		{"Double", "number"},
		{"Boolean", "boolean"},
		{"Bool", "boolean"},
		{"True", "boolean"},
		{"False", "boolean"},
		{"String", "string"},
		{"Message", "#/components/schemas/Message"},
		{"User", "#/components/schemas/User"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := g.convertType(tt.in); got != tt.want {
				t.Errorf("convertType(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestConvertStringSliceToDataType(t *testing.T) {
	g := newTestGen()
	tests := []struct {
		name      string
		in        []string
		wantTypes []string
		wantArray bool
		wantDepth int
	}{
		{"single primitive", []string{"String"}, []string{"String"}, false, 0},
		{"array of primitive", []string{"Array of String"}, []string{"String"}, true, 1},
		{"nested array", []string{"Array of Array of Integer"}, []string{"Integer"}, true, 2},
		{"multi type no array", []string{"Integer", "String"}, []string{"Integer", "String"}, false, 0},
		{
			name:      "multi type each array depth 1",
			in:        []string{"Array of String", "Array of Integer"},
			wantTypes: []string{"String", "Integer"},
			wantArray: true,
			wantDepth: 1,
		},
		{
			name:      "mixed depths keeps max",
			in:        []string{"Array of String", "Array of Array of Integer"},
			wantTypes: []string{"String", "Integer"},
			wantArray: true,
			wantDepth: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.convertStringSliceToDataType(tt.in)
			if !reflect.DeepEqual(got.Types, tt.wantTypes) {
				t.Errorf("Types = %v, want %v", got.Types, tt.wantTypes)
			}
			if got.IsArray != tt.wantArray {
				t.Errorf("IsArray = %v, want %v", got.IsArray, tt.wantArray)
			}
			if got.ArrayDepth != tt.wantDepth {
				t.Errorf("ArrayDepth = %d, want %d", got.ArrayDepth, tt.wantDepth)
			}
		})
	}
}

func TestConvertDataTypeToProperty(t *testing.T) {
	g := newTestGen()

	t.Run("empty types defaults to object", func(t *testing.T) {
		got := g.convertDataTypeToProperty(telegram.DataType{})
		if got.Type != "object" {
			t.Errorf("expected type object, got %+v", got)
		}
	})

	t.Run("single primitive", func(t *testing.T) {
		got := g.convertDataTypeToProperty(telegram.DataType{Types: []string{"String"}})
		if got.Type != "string" || got.Ref != "" {
			t.Errorf("expected {type:string}, got %+v", got)
		}
	})

	t.Run("single ref", func(t *testing.T) {
		got := g.convertDataTypeToProperty(telegram.DataType{Types: []string{"User"}})
		if got.Ref != "#/components/schemas/User" || got.Type != "" {
			t.Errorf("expected {$ref:User}, got %+v", got)
		}
	})

	t.Run("multiple types -> oneOf", func(t *testing.T) {
		got := g.convertDataTypeToProperty(telegram.DataType{Types: []string{"Integer", "String"}})
		if len(got.OneOf) != 2 {
			t.Fatalf("expected oneOf of 2, got %+v", got)
		}
		if got.OneOf[0].Type != "integer" || got.OneOf[1].Type != "string" {
			t.Errorf("unexpected oneOf contents: %+v", got.OneOf)
		}
	})

	t.Run("array of primitive", func(t *testing.T) {
		got := g.convertDataTypeToProperty(telegram.DataType{Types: []string{"String"}, IsArray: true, ArrayDepth: 1})
		if got.Type != "array" || got.Items == nil || got.Items.Type != "string" {
			t.Errorf("expected array of string, got %+v", got)
		}
	})

	t.Run("array of ref", func(t *testing.T) {
		got := g.convertDataTypeToProperty(telegram.DataType{Types: []string{"User"}, IsArray: true, ArrayDepth: 1})
		if got.Type != "array" || got.Items == nil || got.Items.Ref != "#/components/schemas/User" {
			t.Errorf("expected array of $ref User, got %+v", got)
		}
	})

	t.Run("array of multiple types -> items oneOf", func(t *testing.T) {
		got := g.convertDataTypeToProperty(telegram.DataType{Types: []string{"Integer", "String"}, IsArray: true, ArrayDepth: 1})
		if got.Type != "array" || got.Items == nil || len(got.Items.OneOf) != 2 {
			t.Errorf("expected array of oneOf, got %+v", got)
		}
	})

	t.Run("nested array depth 2", func(t *testing.T) {
		got := g.convertDataTypeToProperty(telegram.DataType{Types: []string{"String"}, IsArray: true, ArrayDepth: 2})
		if got.Type != "array" || got.Items == nil || got.Items.Type != "array" {
			t.Fatalf("expected array of array, got %+v", got)
		}
		if got.Items.Items == nil || got.Items.Items.Type != "string" {
			t.Errorf("expected inner items string, got %+v", got.Items.Items)
		}
	})
}

func TestConvertTypesToProperties(t *testing.T) {
	g := newTestGen()
	got := g.convertTypesToProperties([]string{"String", "User"})
	if len(got) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(got))
	}
	if got[0].Type != "string" {
		t.Errorf("expected string property, got %+v", got[0])
	}
	if got[1].Ref != "#/components/schemas/User" {
		t.Errorf("expected $ref User, got %+v", got[1])
	}
}

func TestConvertMethodReturnType(t *testing.T) {
	g := newTestGen()

	t.Run("array return type", func(t *testing.T) {
		got := g.convertMethodReturnType(telegram.ReturnType{Name: "Update", IsArray: true})
		if got.Type != "array" || got.Items == nil || got.Items.Ref != "#/components/schemas/Update" {
			t.Errorf("expected array of $ref Update, got %+v", got)
		}
	})

	tests := []struct {
		name     string
		in       telegram.ReturnType
		wantType string
		wantRef  string
	}{
		{"empty -> boolean", telegram.ReturnType{Name: ""}, "boolean", ""},
		{"integer", telegram.ReturnType{Name: "integer"}, "integer", ""},
		{"boolean", telegram.ReturnType{Name: "boolean"}, "boolean", ""},
		{"String", telegram.ReturnType{Name: "String"}, "string", ""},
		{"True", telegram.ReturnType{Name: "True"}, "boolean", ""},
		{"False", telegram.ReturnType{Name: "False"}, "boolean", ""},
		{"ref type", telegram.ReturnType{Name: "Message"}, "", "#/components/schemas/Message"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.convertMethodReturnType(tt.in)
			if got.Type != tt.wantType || got.Ref != tt.wantRef {
				t.Errorf("got %+v, want type=%q ref=%q", got, tt.wantType, tt.wantRef)
			}
		})
	}
}

func TestDetectUnionTypes(t *testing.T) {
	g := NewWithType(zap.NewNop(), "1.0", map[string]telegram.Type{
		"ChatMember": {
			Name:        "ChatMember",
			Description: "This object contains information about one member of a chat. It can be one of\n- ChatMemberOwner\n- ChatMemberMember\n",
		},
		"User": {
			Name:        "User",
			Description: "This object represents a user.",
			Fields:      []telegram.Field{{Name: "id", Type: []string{"Integer"}}},
		},
	}, nil, "botapi")

	unions := g.detectUnionTypes()
	if _, ok := unions["User"]; ok {
		t.Error("User has fields and must not be detected as a union")
	}
	variants, ok := unions["ChatMember"]
	if !ok {
		t.Fatal("ChatMember should be detected as a union")
	}
	want := []string{"ChatMemberOwner", "ChatMemberMember"}
	if !reflect.DeepEqual(variants, want) {
		t.Errorf("variants = %v, want %v", variants, want)
	}
}
