package cli

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMissingFlagValue = errors.New("flag requires an argument")
	ErrUnknownFlag      = errors.New("unknown command-line flag")
)

type Options struct {
	AutoStart      bool
	ConfigPath     string
	DebugLogging   bool
	DryRun         bool
	StartMinimized bool
}

func ParseArgs(args []string) (*Options, error) {
	opts := &Options{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-a" || arg == "--autostart":
			opts.AutoStart = true
		case arg == "-d" || arg == "--debug":
			opts.DebugLogging = true
		case arg == "--dry-run":
			opts.DryRun = true
		case arg == "-m" || arg == "--minimized":
			opts.StartMinimized = true
		case arg == "-c" || arg == "--config":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return nil, fmt.Errorf("%w: %s requires a path", ErrMissingFlagValue, arg)
			}
			i++
			opts.ConfigPath = args[i]
		case strings.HasPrefix(arg, "--config="):
			opts.ConfigPath = strings.TrimPrefix(arg, "--config=")
		case strings.HasPrefix(arg, "-c="):
			opts.ConfigPath = strings.TrimPrefix(arg, "-c=")
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("%w: %s", ErrUnknownFlag, arg)
			}
		}
	}

	return opts, nil
}
