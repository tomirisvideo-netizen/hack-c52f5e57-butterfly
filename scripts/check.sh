#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
unset OPENAI_API_KEY NVIDIA_API_KEY NEMOTRON_API_KEY NEMOTRON_BASE_URL NEMOTRON_MODEL LOCAL_LLM_BASE_URL LOCAL_LLM_MODEL RIVA_ASR_ENDPOINT ASR_BASE_URL ASR_API_KEY TELEGRAM_BOT_TOKEN TELEGRAM_CHAT_ID
UNFORMATTED=$(gofmt -l cmd internal)
if [ -n "$UNFORMATTED" ]; then
  printf '%s\n' "$UNFORMATTED"
  exit 1
fi
go build ./...
go vet ./...
go test -race ./... -count=1
.venv/bin/python -m pip check
.venv/bin/python -m pytest -q
node --check frontend/app.js
node --test tests/test_frontend.cjs
