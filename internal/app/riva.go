package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"
)

// The HTTP application remains Go. This bounded local subprocess reuses the
// Riva SDK already tested in the Brev assistant container; it is not a server.
func transcribeRiva(ctx context.Context, path, language string) ([]any, error) {
	endpoint := strings.TrimSpace(os.Getenv("RIVA_ASR_ENDPOINT"))
	if host, port, err := net.SplitHostPort(endpoint); err != nil || host == "" || port == "" || strings.ContainsAny(endpoint, "/?#@ \t\n") {
		return nil, providerFailure("Укажите RIVA_ASR_ENDPOINT как имя-хоста:порт.")
	}
	python := strings.TrimSpace(os.Getenv("RIVA_PYTHON"))
	if python == "" {
		python = filepath.Join(".venv", "bin", "python")
		if runtime.GOOS == "windows" {
			python = filepath.Join(".venv", "Scripts", "python.exe")
		}
	}
	source, err := filepath.Abs(path)
	if err != nil {
		return nil, providerFailure("Аудиофайл недоступен.")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, python, "scripts/riva_transcribe.py", "--input", source, "--endpoint", endpoint, "--language", language)
	var output boundedRivaOutput
	cmd.Stdout = &output
	// Never return Python tracebacks (which may contain private paths or audio).
	if err := cmd.Run(); err != nil {
		return nil, providerFailure("Riva ASR не завершил распознавание. Проверьте RIVA_PYTHON, SDK, ffmpeg, язык и доступ к серверу на Brev.")
	}
	return parseRivaOutput(output.Bytes())
}

type boundedRivaOutput struct{ bytes.Buffer }

func (b *boundedRivaOutput) Write(data []byte) (int, error) {
	if b.Len()+len(data) > 8<<20 {
		return 0, fmt.Errorf("Riva response exceeds limit")
	}
	return b.Buffer.Write(data)
}

func parseRivaOutput(raw []byte) ([]any, error) {
	var payload struct {
		Segments []struct {
			Text  string  `json:"text"`
			Start float64 `json:"start"`
			End   float64 `json:"end"`
		} `json:"segments"`
		Diarization bool `json:"diarization"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil || payload.Diarization || len(payload.Segments) == 0 || len(payload.Segments) > 20000 {
		return nil, providerFailure("Riva ASR вернул некорректный транскрипт.")
	}
	segments := make([]any, 0, len(payload.Segments))
	total := 0
	previous := 0.0
	for _, segment := range payload.Segments {
		value := strings.TrimSpace(segment.Text)
		if value == "" {
			continue
		}
		start, err := numeric(segment.Start)
		if err != nil {
			return nil, err
		}
		end, err := numeric(segment.End)
		if err != nil || segment.End < segment.Start || segment.Start < previous {
			return nil, providerFailure("Riva ASR вернул некорректные временные метки.")
		}
		total += utf8.RuneCountInString(value)
		if total > maxTranscriptChars {
			return nil, providerFailure("Результат ASR слишком большой для обработки.")
		}
		previous = segment.Start
		segments = append(segments, map[string]any{"id": fmt.Sprintf("s%d", len(segments)+1), "start": start, "end": end, "speaker": "Спикер не определён", "text": value})
	}
	if len(segments) == 0 {
		return nil, providerFailure("Riva ASR не распознал речь.")
	}
	return segments, nil
}
