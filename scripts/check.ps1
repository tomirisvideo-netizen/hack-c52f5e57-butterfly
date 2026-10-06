$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Invoke-Checked {
    param([string]$Command, [string[]]$Arguments)
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$Command failed with exit code $LASTEXITCODE."
    }
}

$isolatedVariables = @(
    'OPENAI_API_KEY', 'NVIDIA_API_KEY', 'NEMOTRON_API_KEY',
    'NEMOTRON_BASE_URL', 'NEMOTRON_MODEL', 'LOCAL_LLM_BASE_URL', 'LOCAL_LLM_MODEL',
    'RIVA_ASR_ENDPOINT', 'ASR_BASE_URL', 'ASR_API_KEY',
    'TELEGRAM_BOT_TOKEN', 'TELEGRAM_CHAT_ID'
)
$savedEnvironment = @{}
Push-Location (Split-Path -Parent $PSScriptRoot)
try {
    foreach ($name in $isolatedVariables) {
        $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
        [Environment]::SetEnvironmentVariable($name, $null, 'Process')
    }

    $python = Join-Path (Get-Location).Path '.venv/Scripts/python.exe'
    if (-not (Test-Path -LiteralPath $python -PathType Leaf)) {
        throw 'Create the virtual environment first: python -m venv .venv; .venv/Scripts/python.exe -m pip install -r requirements-dev.txt'
    }
    $unformatted = & gofmt -l cmd internal
    if ($LASTEXITCODE -ne 0) { throw 'gofmt failed.' }
    if ($unformatted) { throw "Go files need formatting:`n$($unformatted -join "`n")" }

    Invoke-Checked -Command go -Arguments @('build', './...')
    Invoke-Checked -Command go -Arguments @('vet', './...')
    # The race detector needs a C compiler on Windows; fail if it is unavailable.
    Invoke-Checked -Command go -Arguments @('test', '-race', './...', '-count=1')
    Invoke-Checked -Command $python -Arguments @('-m', 'pip', 'check')
    Invoke-Checked -Command $python -Arguments @('-m', 'pytest', '-q')
    Invoke-Checked -Command node -Arguments @('--check', 'frontend/app.js')
    Invoke-Checked -Command node -Arguments @('--test', 'tests/test_frontend.cjs')
}
finally {
    foreach ($name in $savedEnvironment.Keys) {
        [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process')
    }
    Pop-Location
}
