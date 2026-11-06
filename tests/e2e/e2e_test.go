package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tpsapp/applauncher3/pkg/cli"
	"github.com/tpsapp/applauncher3/pkg/models"
	"github.com/tpsapp/applauncher3/pkg/runner"
	"github.com/tpsapp/applauncher3/pkg/storage"
	"github.com/tpsapp/applauncher3/pkg/tray"
)

func TestE2E_FullLifecycleExecution(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "e2e_config.json")
	markerFile1 := filepath.Join(tmpDir, "task1.done")
	markerFile2 := filepath.Join(tmpDir, "task2.done")

	// 1. Verify CLI argument parsing
	cliOpts, err := cli.ParseArgs([]string{"--config", configPath, "--debug", "--dry-run"})
	if err != nil {
		t.Fatalf("CLI parse failed: %v", err)
	}
	if !cliOpts.DryRun || !cliOpts.DebugLogging {
		t.Fatalf("CLI options unexpected: %+v", cliOpts)
	}

	// 2. Initialize storage manager & seed config
	mgr, err := storage.NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Storage manager init failed: %v", err)
	}

	cfg := models.NewDefaultConfig()
	cfg.ActiveProfile = "E2EProfile"
	cfg.Profiles = []models.Profile{
		{
			Name: "E2EProfile",
			Items: []models.LaunchItem{
				{
					Name:             "Task 1",
					Path:             "/bin/sh",
					Arguments:        []string{"-c", "touch " + markerFile1},
					WorkingDirectory: tmpDir,
					DelaySeconds:     0.05,
					Enabled:          true,
				},
				{
					Name:             "Task 2",
					Path:             "/bin/sh",
					Arguments:        []string{"-c", "touch " + markerFile2},
					WorkingDirectory: tmpDir,
					DelaySeconds:     0.05,
					Enabled:          true,
				},
			},
		},
	}

	if err := mgr.Save(cfg); err != nil {
		t.Fatalf("Config save failed: %v", err)
	}

	// 3. Reload config from disk to verify atomic persistence
	reloaded, err := mgr.Load()
	if err != nil {
		t.Fatalf("Config reload failed: %v", err)
	}
	if len(reloaded.Profiles[0].Items) != 2 {
		t.Fatalf("Expected 2 items in reloaded config")
	}

	// 4. Run real execution engine with OS process spawner
	spawner := runner.NewOSSpawner()
	engine := runner.NewEngine(spawner)

	var recordedEvents []runner.ProgressEvent
	err = engine.Start(context.Background(), reloaded.Profiles[0], false, func(evt runner.ProgressEvent) {
		recordedEvents = append(recordedEvents, evt)
	})

	if err != nil {
		t.Fatalf("Engine start failed: %v", err)
	}

	if engine.GetState() != runner.StateCompleted {
		t.Errorf("Expected StateCompleted, got %s", engine.GetState())
	}

	// Give child processes time to finish writing marker files
	time.Sleep(100 * time.Millisecond)

	// 5. Verify filesystem state changes triggered by spawned children
	if _, err := os.Stat(markerFile1); os.IsNotExist(err) {
		t.Errorf("Marker file 1 was not created by Task 1")
	}
	if _, err := os.Stat(markerFile2); os.IsNotExist(err) {
		t.Errorf("Marker file 2 was not created by Task 2")
	}

	// 6. Test notification dispatch (does not crash or hang)
	tray.Notify("AppLauncher3 Test", "E2E Queue Completed")
}
