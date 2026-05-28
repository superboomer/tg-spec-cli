package commands

import (
	"os"
	"os/exec"
	"testing"
)

func TestExecute(t *testing.T) {
	// With no arguments the root command prints usage and returns nil, so
	// Execute completes without calling os.Exit.
	rootCmd.SetArgs([]string{})
	defer rootCmd.SetArgs(nil)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute() panicked: %v", r)
		}
	}()
	Execute()
}

// TestExecute_ErrorExits verifies that Execute exits non-zero when the root
// command returns an error. Because that path calls os.Exit, it is exercised
// in a re-executed subprocess.
func TestExecute_ErrorExits(t *testing.T) {
	if os.Getenv("BE_EXECUTE_CRASHER") == "1" {
		rootCmd.SetArgs([]string{"this-command-does-not-exist"})
		Execute()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExecute_ErrorExits") //nolint:gosec
	cmd.Env = append(os.Environ(), "BE_EXECUTE_CRASHER=1")
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok && !exitErr.Success() {
		return // expected: non-zero exit from os.Exit(1)
	}
	t.Fatalf("expected Execute() to exit non-zero, got err=%v", err)
}
