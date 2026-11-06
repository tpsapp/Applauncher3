package storage

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/tpsapp/applauncher3/pkg/models"
)

func TestTC_STO_01_AtomicWriteAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom_config.json")

	mgr, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to initialize manager: %v", err)
	}

	cfg := models.NewDefaultConfig()
	cfg.ActiveProfile = "Work"
	cfg.Profiles = append(cfg.Profiles, models.Profile{
		Name: "Work",
		Items: []models.LaunchItem{
			{Name: "IDE", Path: "/usr/bin/code", Enabled: true},
		},
	})

	if err := mgr.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loaded, err := mgr.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.ActiveProfile != "Work" {
		t.Errorf("Expected ActiveProfile 'Work', got '%s'", loaded.ActiveProfile)
	}
	if len(loaded.Profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(loaded.Profiles))
	}
}

func TestTC_STO_02_CustomPathResolution(t *testing.T) {
	custom := "/tmp/test_dir/my_conf.json"
	mgr, err := NewConfigManager(custom)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	abs, _ := filepath.Abs(custom)
	if mgr.ResolvePath() != abs {
		t.Errorf("Expected resolved path '%s', got '%s'", abs, mgr.ResolvePath())
	}
}

func TestTC_STO_03_DefaultCreationWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent", "config.json")

	mgr, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to initialize manager: %v", err)
	}

	loaded, err := mgr.Load()
	if err != nil {
		t.Fatalf("Expected load to create default without error, got: %v", err)
	}

	if loaded.ActiveProfile != "Default" {
		t.Errorf("Expected Default active profile, got '%s'", loaded.ActiveProfile)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected default config.json to be created on disk")
	}
}

func TestTC_STO_04_CorruptJSONQuarantine(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "corrupt_config.json")

	// Write garbage bytes
	if err := os.WriteFile(configPath, []byte("NOT_JSON_<<<CORRUPTED>>>"), 0644); err != nil {
		t.Fatalf("Failed to write corrupt file: %v", err)
	}

	mgr, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	loaded, err := mgr.Load()
	if !errors.Is(err, ErrMalformedConfig) {
		t.Fatalf("Expected ErrMalformedConfig error, got: %v", err)
	}

	if loaded == nil || loaded.ActiveProfile != "Default" {
		t.Errorf("Expected fallback default config returned")
	}

	// Verify original file was quarantined and fresh default created
	entries, _ := os.ReadDir(tmpDir)
	quarantinedFound := false
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			quarantinedFound = true
			break
		}
	}
	if !quarantinedFound {
		t.Errorf("Expected quarantined file in directory")
	}
}

func TestTC_STO_05_ReadOnlyPermissionError(t *testing.T) {
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create readonly dir: %v", err)
	}
	defer func() { _ = os.Chmod(readOnlyDir, 0755) }()

	configPath := filepath.Join(readOnlyDir, "config.json")
	mgr, _ := NewConfigManager(configPath)

	cfg := models.NewDefaultConfig()
	err := mgr.Save(cfg)
	if err == nil {
		t.Fatalf("Expected permission error when saving to read-only directory")
	}
}

func TestTC_STO_06_ConcurrentSavesUnderRaceDetector(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "concurrent_config.json")
	mgr, _ := NewConfigManager(configPath)

	var wg sync.WaitGroup
	workers := 15

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			cfg := models.NewDefaultConfig()
			cfg.Profiles[0].Name = "Default"
			_ = mgr.Save(cfg)
			_, _ = mgr.Load()
		}(i)
	}

	wg.Wait()

	finalConfig, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load failed after concurrent operations: %v", err)
	}
	if finalConfig.ActiveProfile != "Default" {
		t.Errorf("Unexpected active profile: %s", finalConfig.ActiveProfile)
	}
}
