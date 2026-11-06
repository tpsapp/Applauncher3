package models

import (
	"crypto/rand"
	"errors"
	"fmt"
)

var (
	ErrEmptyPath     = errors.New("executable path cannot be empty")
	ErrNegativeDelay = errors.New("delay seconds cannot be negative")
)

type WindowState string

const (
	WindowNormal    WindowState = "Normal"
	WindowMinimized WindowState = "Minimized"
	WindowHidden    WindowState = "Hidden"
)

// GenerateUUIDv4 generates an RFC 4122 compliant version 4 UUID.
func GenerateUUIDv4() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10xxxxxx
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

type LaunchItem struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Path             string      `json:"path"`
	Arguments        []string    `json:"arguments"`
	WorkingDirectory string      `json:"workingDirectory"`
	DelaySeconds     float64     `json:"delaySeconds"`
	LaunchState      WindowState `json:"launchState"`
	Enabled          bool        `json:"enabled"`
}

func (item *LaunchItem) Validate() error {
	if item.ID == "" {
		item.ID = GenerateUUIDv4()
	}
	if item.Path == "" {
		return ErrEmptyPath
	}
	if item.DelaySeconds < 0 {
		return ErrNegativeDelay
	}
	if item.LaunchState == "" {
		item.LaunchState = WindowNormal
	}
	if item.Arguments == nil {
		item.Arguments = []string{}
	}
	return nil
}
