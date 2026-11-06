package models

import "errors"

var (
	ErrDuplicateProfileName  = errors.New("duplicate profile name")
	ErrActiveProfileNotFound = errors.New("active profile not found in profiles list")
	ErrNoProfilesDefined     = errors.New("configuration must contain at least one profile")
)

type AppSettings struct {
	StartMinimized  bool   `json:"startMinimized"`
	CloseToTray     bool   `json:"closeToTray"`
	ConfirmOnCancel bool   `json:"confirmOnCancel"`
	Theme           string `json:"theme"` // "system", "dark", "light"
	DebugLogging    bool   `json:"debugLogging"`
}

type LauncherConfig struct {
	Version       int         `json:"version"`
	ActiveProfile string      `json:"activeProfile"`
	Settings      AppSettings `json:"settings"`
	Profiles      []Profile   `json:"profiles"`
}

func NewDefaultConfig() *LauncherConfig {
	return &LauncherConfig{
		Version:       1,
		ActiveProfile: "Default",
		Settings: AppSettings{
			StartMinimized:  false,
			CloseToTray:     true,
			ConfirmOnCancel: true,
			Theme:           "system",
			DebugLogging:    false,
		},
		Profiles: []Profile{
			{
				Name:  "Default",
				Items: []LaunchItem{},
			},
		},
	}
}

func (c *LauncherConfig) Validate() error {
	if c.Version < 1 {
		c.Version = 1
	}
	if len(c.Profiles) == 0 {
		return ErrNoProfilesDefined
	}

	profileNames := make(map[string]bool)
	activeFound := false

	for i := range c.Profiles {
		p := &c.Profiles[i]
		if err := p.Validate(); err != nil {
			return err
		}
		if profileNames[p.Name] {
			return ErrDuplicateProfileName
		}
		profileNames[p.Name] = true
		if p.Name == c.ActiveProfile {
			activeFound = true
		}
	}

	if !activeFound {
		return ErrActiveProfileNotFound
	}

	if c.Settings.Theme == "" {
		c.Settings.Theme = "system"
	}

	return nil
}

// GetActiveProfile returns a pointer to the active profile or nil if not found.
func (c *LauncherConfig) GetActiveProfile() *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Name == c.ActiveProfile {
			return &c.Profiles[i]
		}
	}
	return nil
}
