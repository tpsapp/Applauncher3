package runner

import (
	"errors"
	"os"

	"github.com/tpsapp/applauncher3/pkg/models"
)

var (
	ErrDirectoryNotFound = errors.New("working directory does not exist")
	ErrPathNotFound      = errors.New("executable path not found")
)

type SpawnResult struct {
	PID  int
	Wait func() error
}

type ProcessSpawner interface {
	Spawn(item *models.LaunchItem) (*SpawnResult, error)
}

func checkWorkingDirectory(dir string) error {
	if dir == "" {
		return nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrDirectoryNotFound
		}
		return err
	}
	if !info.IsDir() {
		return ErrDirectoryNotFound
	}
	return nil
}
