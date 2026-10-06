# Butterfly — протоколы встреч

Go-приложение: загрузка записи или текста, поручения с цитатами, ручная проверка,
экспорт PDF/DOCX/JSON. Интерфейс выполнен по предоставленному дизайну.
Готовые HackAlem MCP и Fly demo подключаются как отдельные процессы.

## Запуск

Требуются Go 1.27+, Python 3.10+ для локального экспорта.

### Windows (PowerShell)

```powershell
python -m venv .venv
.\.venv\Scripts\python.exe -m pip install -r requirements-dev.txt
if (!(Test-Path .env)) { Copy-Item .env.example .env }
.\scripts\run.ps1
```

### Linux / macOS

```bash
python3 -m venv .venv
.venv/bin/pip install -r requirements.txt
cp .env.example .env
./scripts/run.sh
```

Открыть `http://127.0.0.1:8000`, нажать «Демо». Данные явно помечены синтетическими.
Для готовых MCP/Fly указать `HACKALEM_HOME` — каталог установленного HackAlem
с `bin/hackalem-mcp` и `bin/hackalem-demo-api` (на Windows — файлы `.exe`). Через MCP выполняются маршрутизация,
анонимный отчёт проверки, обратная связь Fly и явная отправка статуса в Telegram.

Основной backend написан на Go. Python используется только для PDF/DOCX и
совместимости с установленным Riva SDK, а не как HTTP-сервер.

Если `BUTTERFLY_EXPORT_PYTHON` и `RIVA_PYTHON` пусты, приложение автоматически
использует Python из `.venv/Scripts` на Windows и `.venv/bin` на Linux/macOS.
При обновлении старого `.env` удалите значения `.venv/bin/python` / `python3`
из этих двух переменных, если они не соответствуют вашей системе. Для Riva
нужен отдельный установленный SDK и доступный сервер; обычные зависимости
экспорта его не устанавливают.

HackAlem подключается из внешнего каталога и не обновляется командой `pip install`.
После обновления его исходников пересоберите **оба** исполняемых файла: старый
`hackalem-demo-api` может отвечать на healthcheck, но не содержать `/mb-3d`.
Butterfly проверяет совместимость страницы перед показом Fly.

## Nemotron на Brev

Модели уже запущены отдельной сессией. Следовать
[актуальной инструкции интеграции](docs/deployment/nemotron-integration-handoff.md)
и [параметрам и результатам VM](docs/deployment/runtime-handoff.json).
ASR — Riva gRPC, LLM — локальный Nemotron HTTP API. Нельзя подставлять адрес
`127.0.0.1` VM в конфигурацию сервера, работающего на другой машине.

В репозитории проверены синтетический сценарий, экспорт, права доступа и
реальные процессы MCP/Fly. Другой сессией подтверждён ASR→LLM на 30 секундах
русского аудио. Полный путь нашего приложения через Brev, публичный HTTPS backend,
Vercel, диаризация и качество KK/mixed пока не подтверждены.

Для отдельного frontend: в настройках указать HTTPS backend; на backend задать
точный `BUTTERFLY_CORS_ORIGINS`, `BUTTERFLY_COOKIE_SAMESITE=none` и
`BUTTERFLY_SECURE_COOKIE=true`. Same-origin размещение проще для браузерных cookies.
Telegram требует серверные `TELEGRAM_BOT_TOKEN` и `TELEGRAM_CHAT_ID` и отправляет
только статус по кнопке пользователя; тексты встреч в сообщение не включаются.

## Проверки

Полная проверка на Windows: `.\scripts\check.ps1`; на Linux/macOS:
`sh scripts/check.sh`. Включает сборку, vet, Go race detector, Python-экспорт,
совместимость зависимостей и регрессионные тесты интерфейса. Для Go race detector
на Windows нужен доступный C-компилятор (например, GCC).

```bash
go test ./...
go vet ./...
.venv/bin/pip install -r requirements-dev.txt
.venv/bin/python -m pytest -q
node --check frontend/app.js
node --test tests/test_frontend.cjs
# При запущенном сервере, без отправки сообщений в Telegram:
node scripts/smoke.mjs
```

Исходные аудио, секреты, runtime и протоколы пользователей не коммитятся.
[Архитектура](docs/architecture.md), [этапы и доказательства](docs/stage-evidence.md),
[исследование](docs/research/theme-research-meeting-minutes.md).

## Docker и eval

```bash
docker compose up --build -d backend
node scripts/eval.mjs --base http://127.0.0.1:8000
```

[Docker Compose / Brev profile](docs/deployment/docker.md) ·
[Что измеряет eval](docs/evaluation.md) · локальный CI: `sh scripts/check.sh`.
Regression eval: 15 passed, 0 failed. Docker-файлы подготовлены; запуск контейнера
здесь не проверен, поскольку Docker CLI отсутствует.
