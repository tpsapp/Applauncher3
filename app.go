package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/tpsapp/applauncher3/pkg/cli"
	"github.com/tpsapp/applauncher3/pkg/logger"
	"github.com/tpsapp/applauncher3/pkg/models"
	"github.com/tpsapp/applauncher3/pkg/runner"
	"github.com/tpsapp/applauncher3/pkg/storage"
)

type App struct {
	ctx        context.Context
	configMgr  *storage.ConfigManager
	engine     *runner.Engine
	cliOptions *cli.Options
	mu         sync.Mutex
}

func NewApp(cliOptions *cli.Options) (*App, error) {
	customPath := ""
	if cliOptions != nil {
		customPath = cliOptions.ConfigPath
	}

	mgr, err := storage.NewConfigManager(customPath)
	if err != nil {
		return nil, err
	}

	appEngine := runner.NewEngine(nil)

	return &App{
		configMgr:  mgr,
		engine:     appEngine,
		cliOptions: cliOptions,
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load configuration
	cfg, err := a.configMgr.Load()
	if err == nil {
		enableLogging := cfg.Settings.DebugLogging
		if a.cliOptions != nil && a.cliOptions.DebugLogging {
			enableLogging = true
		}
		_ = logger.InitLogger("launcher-debug.log", enableLogging)
		logger.Info("AppLauncher3 started. Configuration loaded from %s", a.configMgr.ResolvePath())
	}

	// Handle autostart CLI flag
	if a.cliOptions != nil && a.cliOptions.AutoStart {
		go func() {
			// Allow frontend UI to mount and register listeners
			time.Sleep(500 * time.Millisecond)
			logger.Info("AutoStart flag active. Beginning queue execution...")
			_ = a.StartQueue("")
		}()
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.engine != nil {
		a.engine.Cancel()
	}
	_ = logger.Close()
}

// beforeClose intercepts the window close event and checks CloseToTray preference.
func (a *App) beforeClose(ctx context.Context) bool {
	cfg, err := a.configMgr.Load()
	if err != nil {
		return false // Permit close
	}

	if cfg.Settings.CloseToTray {
		logger.Info("Window close requested. Minimizing to background (CloseToTray=true)...")
		if ctx != nil && ctx.Value("frontend") != nil {
			runtime.WindowHide(ctx)
		}
		return true // Prevent window termination, keep background running
	}

	logger.Info("Window close requested. Exiting application immediately (CloseToTray=false)...")
	return false // Permit window termination and app exit
}

// GetConfig returns the loaded configuration.
func (a *App) GetConfig() (*models.LauncherConfig, error) {
	return a.configMgr.Load()
}

// SaveConfig validates and saves the configuration to disk.
func (a *App) SaveConfig(cfg *models.LauncherConfig) error {
	logger.Info("Saving configuration...")
	return a.configMgr.Save(cfg)
}

// StartQueue launches execution of the specified profile (or active profile if empty).
func (a *App) StartQueue(profileName string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	cfg, err := a.configMgr.Load()
	if err != nil {
		return err
	}

	targetProfile := cfg.GetActiveProfile()
	if profileName != "" {
		for i := range cfg.Profiles {
			if cfg.Profiles[i].Name == profileName {
				targetProfile = &cfg.Profiles[i]
				break
			}
		}
	}

	if targetProfile == nil {
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	dryRun := false
	if a.cliOptions != nil && a.cliOptions.DryRun {
		dryRun = true
	}

	logger.Info("Starting queue for profile '%s' (Items: %d, DryRun: %v)", targetProfile.Name, len(targetProfile.Items), dryRun)

	go func() {
		err := a.engine.Start(context.Background(), *targetProfile, dryRun, func(evt runner.ProgressEvent) {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "queue:progress", evt)
			}
			if evt.ItemStatus == "Failed" {
				logger.Error("Item '%s' failed: %s", evt.CurrentItem.Name, evt.ErrorMessage)
			}
		})
		if err != nil {
			logger.Error("Queue execution error: %v", err)
		} else {
			logger.Info("Queue execution completed.")
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "queue:completed", targetProfile.Name)
			}
		}
	}()

	return nil
}

func (a *App) PauseQueue() {
	logger.Info("Pausing queue...")
	a.engine.Pause()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "queue:state_changed", "Paused")
	}
}

func (a *App) ResumeQueue() {
	logger.Info("Resuming queue...")
	a.engine.Resume()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "queue:state_changed", "Running")
	}
}

func (a *App) CancelQueue() {
	logger.Info("Cancelling queue...")
	a.engine.Cancel()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "queue:state_changed", "Cancelled")
	}
}

func (a *App) GetQueueState() string {
	return string(a.engine.GetState())
}

// SelectExecutableDialog opens a native file dialog for selecting an executable or script.
func (a *App) SelectExecutableDialog() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("context not initialized")
	}

	options := runtime.OpenDialogOptions{
		Title: "Select Application or Script to Launch",
		Filters: []runtime.FileFilter{
			{DisplayName: "All Executables & Scripts", Pattern: "*"},
			{DisplayName: "Applications (*.exe;*.app;*.desktop)", Pattern: "*.exe;*.app;*.desktop"},
			{DisplayName: "Scripts (*.sh;*.bat;*.cmd;*.ps1)", Pattern: "*.sh;*.bat;*.cmd;*.ps1"},
		},
	}

	selected, err := runtime.OpenFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	return selected, nil
}

// SelectDirectoryDialog opens a native folder picker for selecting working directories.
func (a *App) SelectDirectoryDialog() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("context not initialized")
	}

	options := runtime.OpenDialogOptions{
		Title: "Select Working Directory",
	}

	return runtime.OpenDirectoryDialog(a.ctx, options)
}

type ItemMetadata struct {
	Name             string `json:"name"`
	WorkingDirectory string `json:"workingDirectory"`
}

// InferItemMetadata returns friendly name and directory given an executable path.
func (a *App) InferItemMetadata(execPath string) ItemMetadata {
	if execPath == "" {
		return ItemMetadata{}
	}
	base := filepath.Base(execPath)
	ext := filepath.Ext(base)
	return ItemMetadata{
		Name:             strings.TrimSuffix(base, ext),
		WorkingDirectory: filepath.Dir(execPath),
	}
}
