package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/tpsapp/applauncher3/pkg/models"
	"github.com/tpsapp/applauncher3/pkg/storage"
)

func TestInferItemMetadata(t *testing.T) {
	app := &App{}

	testCases := []struct {
		input       string
		expectedName string
		expectedDir  string
	}{
		{
			input:       "/home/tpsapp/.local/bin/dicemachine",
			expectedName: "dicemachine",
			expectedDir:  "/home/tpsapp/.local/bin",
		},
		{
			input:       "/usr/bin/google-chrome-stable",
			expectedName: "google-chrome-stable",
			expectedDir:  "/usr/bin",
		},
		{
			input:       "/opt/scripts/deploy.sh",
			expectedName: "deploy",
			expectedDir:  "/opt/scripts",
		},
	}

	for _, tc := range testCases {
		meta := app.InferItemMetadata(tc.input)
		if meta.Name != tc.expectedName {
			t.Errorf("For input %s, expected Name '%s', got '%s'", tc.input, tc.expectedName, meta.Name)
		}
		if meta.WorkingDirectory != tc.expectedDir {
			t.Errorf("For input %s, expected WorkingDirectory '%s', got '%s'", tc.input, tc.expectedDir, meta.WorkingDirectory)
		}
	}
}

func TestBeforeClose_RespectsCloseToTray(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	mgr, err := storage.NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create ConfigManager: %v", err)
	}

	app := &App{configMgr: mgr}
	ctx := context.Background()

	// 1. With CloseToTray = false, beforeClose must return false (exit immediately)
	cfg := models.NewDefaultConfig()
	cfg.Settings.CloseToTray = false
	if err := mgr.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	prevent := app.beforeClose(ctx)
	if prevent {
		t.Errorf("Expected beforeClose to return false (immediate exit) when CloseToTray=false, got true")
	}

	// 2. With CloseToTray = true, beforeClose must return true (prevent termination)
	cfg.Settings.CloseToTray = true
	if err := mgr.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	prevent = app.beforeClose(ctx)
	if !prevent {
		t.Errorf("Expected beforeClose to return true (prevent exit) when CloseToTray=true, got false")
	}
}
