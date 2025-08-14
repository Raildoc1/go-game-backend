package service

import (
	"flag"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig[TConfig any](defaultPath string, debugWriter io.Writer, onCloseError func(error)) (*TConfig, error) {
	var configPath string
	flag.StringVar(&configPath, "config", defaultPath, "config in YAML format")
	flag.Parse()

	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", configPath, err)
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
