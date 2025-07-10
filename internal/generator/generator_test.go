package generator

import (
	"testing"

	"github.com/superboomer/tg-spec-cli/internal/telegram"
	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	log := zaptest.NewLogger(t)
	gen := New(log, "1.0", map[string]telegram.Type{}, []telegram.Method{})
	if gen == nil {
		t.Error("New() returned nil")
	}
}

func TestNewWithType(t *testing.T) {
	log := zaptest.NewLogger(t)
	gen := NewWithType(log, "1.0", map[string]telegram.Type{}, []telegram.Method{}, "gateway")
	if gen == nil {
		t.Error("NewWithType() returned nil")
		return
	}
	if gen.typeFlag != "gateway" {
		t.Errorf("typeFlag = %v, want gateway", gen.typeFlag)
	}
}

func TestGenerator_Generate(t *testing.T) {
	log := zaptest.NewLogger(t)
	gen := NewWithType(log, "1.0", map[string]telegram.Type{}, []telegram.Method{}, "gateway")
	_, err := gen.Generate()
	if err != nil {
		t.Logf("Generate() error: %v", err)
	}
}
