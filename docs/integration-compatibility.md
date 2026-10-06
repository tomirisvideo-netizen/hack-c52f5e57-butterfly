# HackAlem integration compatibility

Verified on Windows on 2026-10-07 against upstream commit
`694c890b72b7a9218dc8cbec0a05e99ebf918217` in
[Danonika/hackathon-plugin](https://github.com/Danonika/hackathon-plugin/commit/694c890b72b7a9218dc8cbec0a05e99ebf918217).
The commit title says `v0.16.100`; its root plugin manifest and README still report
`0.16.99`, so the commit SHA identifies the verified runtime precisely.

Both Windows executables were built from a separate private checkout. The global
HackAlem installation was not changed, and upstream source is not vendored into
Butterfly. Set `HACKALEM_HOME` to the checkout containing the matching
`bin/hackalem-mcp.exe` and `bin/hackalem-demo-api.exe`. Linux uses the same binary
names without the `.exe` suffix. Butterfly creates a separate runtime data root
under its own data directory.

## Compatibility fixes

- Closing integrations cancels their private context, stops the event worker,
  and reaps both child processes. Canceling the owner context also stops an idle
  MCP process and the Fly process.
- Runtime health and Telegram availability reflect closed or exited processes.
  MCP's `isError` flag is honored even if a returned envelope claims success.
- Fly startup checks `/mb-3d` as well as `/healthz`. The previously installed
  Windows demo executable, built from `8824cc34febf`, returned 200 for health and
  404 for the required page. Such a binary now produces a compatibility error
  instead of advertising an available visualization.
- The default export interpreter is `.venv/Scripts/python.exe` on Windows and
  `.venv/bin/python` on other platforms. `BUTTERFLY_EXPORT_PYTHON` continues to
  override the default. Relative `HACKALEM_HOME` paths are resolved before child
  processes change their working directory.

## Verification

The focused regressions first reproduced the worker leak, idle MCP child leak,
stale health result, Windows interpreter failure, and ignored MCP `isError` flag.
The package tests and `go vet` passed after the fixes.

The installed-runtime tests are opt-in, because the repository does not bundle
HackAlem executables. To run them against a compatible private checkout:

```powershell
$env:HACKALEM_TEST_HOME = 'C:\path\to\private\hackalem-checkout'
$env:HACKALEM_MCP_TEST_BINARY = Join-Path $env:HACKALEM_TEST_HOME 'bin\hackalem-mcp.exe'
go test -race ./internal/integrations ./internal/mcpbridge -count=1
```

This command passed with the verified commit above. It checks initialization,
`system.context`, `system.capabilities`, `routing.fly_recommend`, `reports.create`,
and `routing.fly_feedback`, using only synthetic metadata and a temporary private
data root. It also checks the authenticated Fly page and event routes, the local
Three.js import, and shutdown. Unit tests cover notification ownership, review
gating, duplicate suppression, and uncertain delivery with a mock MCP caller.

No real Telegram messages, meeting recordings, or live ASR/LLM requests are part
of this verification. The tests do not claim that hosted model credentials or
external deployment configuration work.
