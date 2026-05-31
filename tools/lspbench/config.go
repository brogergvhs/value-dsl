package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func configPaths(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read config directory: %w", err)
	}
	paths := []string(nil)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		paths = append(paths, filepath.Join(path, entry.Name()))
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("no JSON configs found in %q", path)
	}
	return paths, nil
}

func loadConfig(path string) (config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return config{}, fmt.Errorf("parse config: %w", err)
	}
	base := filepath.Dir(path)
	if cfg.Name == "" {
		cfg.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if len(cfg.Command) == 0 {
		return config{}, errors.New("config command is required")
	}
	if cfg.Root == "" || cfg.File == "" || cfg.LanguageID == "" {
		return config{}, errors.New("config root, file, and languageId are required")
	}
	cfg.Root = resolvePath(base, cfg.Root)
	cfg.File = resolvePath(base, cfg.File)
	if cfg.TimeoutMS == 0 {
		cfg.TimeoutMS = 10000
	}
	return cfg, nil
}

func resolvePath(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Clean(filepath.Join(base, path))
}

func (cfg config) waitDiagnostics() bool {
	return cfg.WaitDiagnostics == nil || *cfg.WaitDiagnostics
}
