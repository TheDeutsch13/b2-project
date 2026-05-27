param(
    [string]$Link = "/catalog"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$nodeScript = Join-Path $PSScriptRoot "sync-carousel.mjs"

if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Error "Node.js не найден. Установите Node.js для синхронизации карусели."
}

& node $nodeScript $Link
