package integrations

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("BUTTERFLY_EXPORT_TEST_HELPER") == "1" {
		_, err := io.Copy(os.Stdout, os.Stdin)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestExportUsesPlatformVirtualEnvironment(t *testing.T) {
	t.Setenv("BUTTERFLY_EXPORT_TEST_HELPER", "1")
	t.Setenv("BUTTERFLY_EXPORT_PYTHON", "")
	t.Chdir(t.TempDir())
	python := filepath.Join(".venv", "bin", "python")
	if runtime.GOOS == "windows" {
		python = filepath.Join(".venv", "Scripts", "python.exe")
	}
	if err := os.MkdirAll(filepath.Dir(python), 0700); err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(python, data, 0700); err != nil {
		t.Fatal(err)
	}
	output, err := Export(map[string]any{"title": "Synthetic export"}, "docx")
	if err != nil {
		t.Fatalf("default virtual environment export failed: %v", err)
	}
	var got map[string]any
	if json.Unmarshal(output, &got) != nil || got["title"] != "Synthetic export" {
		t.Fatalf("helper did not receive export input: %q", output)
	}
}

func TestExportHonorsConfiguredInterpreter(t *testing.T) {
	t.Setenv("BUTTERFLY_EXPORT_TEST_HELPER", "1")
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("BUTTERFLY_EXPORT_PYTHON", binary)
	output, err := Export(map[string]any{"title": "Configured helper"}, "pdf")
	if err != nil || !json.Valid(output) {
		t.Fatalf("configured helper failed: %q, %v", output, err)
	}
}
