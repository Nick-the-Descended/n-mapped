// Package config resolves cross-platform paths for n-mapped's persistent data.
package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// DataDir returns the directory where SQLite, scan outputs, and settings live.
// Honors $N_MAPPED_DATA_DIR override when set.
func DataDir() (string, error) {
	if override := os.Getenv("N_MAPPED_DATA_DIR"); override != "" {
		return ensure(override)
	}
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return ensure(filepath.Join(home, "Library", "Application Support", "n-mapped"))
	default:
		base := os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, ".local", "share")
		}
		return ensure(filepath.Join(base, "n-mapped"))
	}
}

func ensure(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
