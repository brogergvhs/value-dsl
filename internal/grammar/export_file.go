package grammar

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteJSONFile(path string) error {
	data, err := ExportJSON()
	if err != nil {
		return fmt.Errorf("export grammar JSON: %w", err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write grammar spec: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("commit grammar spec: %w", err)
	}
	return nil
}
