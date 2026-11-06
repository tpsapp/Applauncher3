package cli

import (
	"errors"
	"testing"
)

func TestTC_CLI_01_HappyPath_LongFlags(t *testing.T) {
	args := []string{"--autostart", "--config", "/tmp/test.json", "--debug", "--dry-run"}
	opts, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}

	if !opts.AutoStart {
		t.Errorf("Expected AutoStart true")
	}
	if opts.ConfigPath != "/tmp/test.json" {
		t.Errorf("Expected ConfigPath '/tmp/test.json', got '%s'", opts.ConfigPath)
	}
	if !opts.DebugLogging {
		t.Errorf("Expected DebugLogging true")
	}
	if !opts.DryRun {
		t.Errorf("Expected DryRun true")
	}
}

func TestTC_CLI_02_ShortAliases(t *testing.T) {
	args := []string{"-a", "-c", "/tmp/app.json", "-d", "-m"}
	opts, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}

	if !opts.AutoStart || !opts.DebugLogging || !opts.StartMinimized {
		t.Errorf("Short boolean flags not set properly: %+v", opts)
	}
	if opts.ConfigPath != "/tmp/app.json" {
		t.Errorf("Expected ConfigPath '/tmp/app.json', got '%s'", opts.ConfigPath)
	}
}

func TestTC_CLI_03_MissingFlagArg(t *testing.T) {
	args := []string{"--config"}
	_, err := ParseArgs(args)
	if !errors.Is(err, ErrMissingFlagValue) {
		t.Fatalf("Expected ErrMissingFlagValue, got: %v", err)
	}
}

func TestTC_CLI_04_UnknownFlag(t *testing.T) {
	args := []string{"--unrecognized-flag-xyz"}
	_, err := ParseArgs(args)
	if !errors.Is(err, ErrUnknownFlag) {
		t.Fatalf("Expected ErrUnknownFlag, got: %v", err)
	}
}
