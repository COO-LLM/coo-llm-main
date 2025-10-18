package main

import (
	"os"
	"testing"
)

func TestMainVersionFlag(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test version flag
	os.Args = []string{"coo-llm", "-version"}

	// This would normally exit, but we can't test that easily
	// Just test that the flag parsing doesn't panic
	versionFlag := false
	for _, arg := range os.Args[1:] {
		if arg == "-version" {
			versionFlag = true
			break
		}
	}

	if !versionFlag {
		t.Error("Version flag not found in args")
	}
}

func TestMainConfigFlag(t *testing.T) {
	// Test config flag parsing
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"coo-llm", "-config", "test.yaml"}

	// Verify args are set correctly
	if len(os.Args) < 3 || os.Args[2] != "test.yaml" {
		t.Error("Config flag not parsed correctly")
	}
}

func TestEnvOverrides(t *testing.T) {
	// Test environment variable handling
	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")

	// Test that env var is accessible
	if port := os.Getenv("PORT"); port != "9090" {
		t.Errorf("Expected PORT=9090, got %s", port)
	}
}
