package service

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads a YAML configuration file into the provided type. The
// defaultPath is used when the `-config` flag is not supplied. If debugWriter
// is non-nil the loaded configuration will be written to it. The onCloseError
// callback is invoked when closing the configuration file fails.
func LoadConfig[TConfig any](defaultPath string, debugWriter io.Writer, onCloseError func(error)) (*TConfig, error) {
	var configPath string
	flag.StringVar(&configPath, "config", defaultPath, "config in YAML format")
	flag.Parse()

	cleanConfigPath, err := validateAndCleanPath(".", configPath)
	if err != nil {
		return nil, fmt.Errorf("error validating path: %w", err)
	}

	// #nosec G304 -- safePath validated against rootDir
	f, err := os.Open(cleanConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", cleanConfigPath, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			if onCloseError != nil {
				onCloseError(err)
			}
		}
	}(f)
	var cfg TConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse default.yaml: %w", err)
	}
	if debugWriter != nil {
		err := yaml.NewEncoder(debugWriter).Encode(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to output config: %w", err)
		}
	}
	return &cfg, nil
}

// validateAndCleanPath ensures that 'path' is inside 'root' (or exactly 'root').
// Both root and path are resolved to absolute, cleaned paths.
// Fails if the target is outside root, even via symlink.
func validateAndCleanPath(root, path string) (string, error) {
	absRoot, err := filepath.EvalSymlinks(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("invalid root path: %w", err)
	}
	absRoot, err = filepath.Abs(absRoot)
	if err != nil {
		return "", fmt.Errorf("invalid root path: %w", err)
	}

	absPath, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("invalid target path: %w", err)
	}
	absPath, err = filepath.Abs(absPath)
	if err != nil {
		return "", fmt.Errorf("invalid target path: %w", err)
	}

	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot) {
		return "", fmt.Errorf("path %q is outside allowed root %q", absPath, absRoot)
	}

	return absPath, nil
}
