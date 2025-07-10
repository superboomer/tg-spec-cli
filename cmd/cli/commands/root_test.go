package commands

import (
	"testing"
)

func TestRootCmd(t *testing.T) {
	if rootCmd.Use != "tg-bot-api-gen" {
		t.Errorf("rootCmd.Use = %v, want 'tg-bot-api-gen'", rootCmd.Use)
	}
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
}
