package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tpsapp/applauncher3/pkg/models"
)

var (
	ErrMalformedConfig = errors.New("configuration file is malformed or corrupted")
)

type ConfigManager struct {
	customPath   string
	resolvedPath string
	mu           sync.RWMutex
}

func NewConfigManager(customPath string) (*ConfigManager, error) {
	mgr := &ConfigManager{
		customPath: customPath,
	}
	resolved, err := mgr.resolveConfigPath()
	if err != nil {
		return nil, err
	}
	mgr.resolvedPath = resolved
	return mgr, nil
}

// ResolvePath returns the active configuration file path.
func (m *ConfigManager) ResolvePath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.resolvedPath
}

func (m *ConfigManager) resolveConfigPath() (string, error) {
	if m.customPath != "" {
		return filepath.Abs(m.customPath)
	}

	// Check portable mode: if config.json exists beside executable or in CWD
	if _, err := os.Stat("config.json"); err == nil {
		return filepath.Abs("config.json")
	}

	// Standard OS user config directory
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		userConfigDir = "."
	}

	appDir := filepath.Join(userConfigDir, "applauncher3")
	return filepath.Join(appDir, "config.json"), nil
}

// Load loads and validates the configuration from disk.
// If the file does not exist, a default configuration is initialized and saved.
func (m *ConfigManager) Load() (*models.LauncherConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.resolvedPath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultCfg := models.NewDefaultConfig()
		if err := m.saveInternal(defaultCfg); err != nil {
			return defaultCfg, nil
		}
		return defaultCfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg models.LauncherConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		// Quarantine corrupted configuration file
		backupPath := fmt.Sprintf("%s.corrupt.%d", path, time.Now().Unix())
		_ = os.Rename(path, backupPath)
		defaultCfg := models.NewDefaultConfig()
		_ = m.saveInternal(defaultCfg)
		return defaultCfg, fmt.Errorf("%w: quarantined to %s", ErrMalformedConfig, backupPath)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save atomically writes the configuration to disk.
func (m *ConfigManager) Save(cfg *models.LauncherConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveInternal(cfg)
}

func (m *ConfigManager) saveInternal(cfg *models.LauncherConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	dir := filepath.Dir(m.resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := fmt.Sprintf("%s.tmp.%d", m.resolvedPath, os.Getpid())
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, m.resolvedPath)
}
