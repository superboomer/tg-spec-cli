package app

import (
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	log := zaptest.NewLogger(t)
	app := New(log, "http://example.com", "output.json")
	if app == nil {
		t.Error("New() returned nil")
	}
}

func TestNewWithType(t *testing.T) {
	log := zaptest.NewLogger(t)
	app := NewWithType(log, "http://example.com", "output.json", "botapi")
	if app == nil {
		t.Error("NewWithType() returned nil")
		return
	}
	if app.typeFlag != "botapi" {
		t.Errorf("typeFlag = %v, want botapi", app.typeFlag)
	}
}

func TestApp_Run_UnsupportedType(t *testing.T) {
	log := zaptest.NewLogger(t)
	app := NewWithType(log, "http://example.com", "output.json", "invalidtype")
	err := app.Run()
	if err == nil {
		t.Error("Run() with unsupported type should return error")
	}
}

// Note: Integration tests for Run() with real HTTP requests are not included here to avoid network dependency.
