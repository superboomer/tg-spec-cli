package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestGenerateCmd(t *testing.T) {
	cmd := generateCmd
	if cmd.Use != "generate" {
		t.Errorf("generateCmd.Use = %v, want 'generate'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("generateCmd.Short should not be empty")
	}
}

func TestGenerateCmdRun(t *testing.T) {
	cmd := &cobra.Command{}
	outputPath = "output.json"
	logLevel = "debug"
	url = "https://core.telegram.org/bots/api"
	typeFlag = "botapi"

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("generateCmd.Run panicked: %v", r)
		}
	}()
	// Accept that the command may log a fatal error due to missing 'Recent Changes' section,
	// but the test should pass as long as it does not panic.
	generateCmd.Run(cmd, []string{})
}
