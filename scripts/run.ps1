$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

Push-Location (Split-Path -Parent $PSScriptRoot)
try {
    New-Item -ItemType Directory -Path bin -Force | Out-Null
    & go build -o bin/butterfly.exe ./cmd/butterfly
    if ($LASTEXITCODE -ne 0) { throw "Go build failed with exit code $LASTEXITCODE." }
    & ./bin/butterfly.exe
    if ($LASTEXITCODE -ne 0) { throw "Butterfly exited with code $LASTEXITCODE." }
}
finally {
    Pop-Location
}
