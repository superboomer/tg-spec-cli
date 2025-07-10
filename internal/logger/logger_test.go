package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	log, err := New("debug")
	if err != nil {
		t.Errorf("New() error = %v", err)
	}
	if log == nil {
		t.Error("New() returned nil logger")
	}

	_, err = New("invalid-level")
	if err == nil {
		t.Error("New() with invalid level should return error")
	}
}

func TestNewSilent(t *testing.T) {
	log, err := New("silent")
	if err != nil {
		t.Errorf("New('silent') error = %v", err)
	}
	if log == nil {
		t.Error("New('silent') returned nil logger")
	}
	// Should not panic or output anything
	log.Info("should not be visible")
}
