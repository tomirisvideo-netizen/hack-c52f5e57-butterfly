package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func Export(meeting map[string]any, format string) ([]byte, error) {
	if format != "pdf" && format != "docx" && format != "json" {
		return nil, errors.New("unsupported format")
	}
	input, err := json.Marshal(meeting)
	if err != nil {
		return nil, err
	}
	python := os.Getenv("BUTTERFLY_EXPORT_PYTHON")
	if python == "" {
		python = filepath.Join(".venv", "bin", "python")
		if runtime.GOOS == "windows" {
			python = filepath.Join(".venv", "Scripts", "python.exe")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, python, "scripts/export_protocol.py", format)
	cmd.Stdin = bytes.NewReader(input)
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.New("local export helper unavailable; check Python export dependencies")
	}
	return output, nil
}
