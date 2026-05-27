# Use when auth migrate fails with "users already exists" or "role already exists"
$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$authPath = Join-Path $root "migrations\auth"
$container = "b2-postgres"

$network = docker inspect -f '{{range $k, $v := .NetworkSettings.Networks}}{{$k}}{{end}}' $container 2>$null
if (-not $network) {
    Write-Host "Container $container is not running."
    exit 1
}

$authDbUrl = "postgres://b2user:b2password@postgres:5432/b2db?sslmode=disable&x-migrations-table=auth_schema_migrations"
$absPath = (Resolve-Path $authPath).Path
$volume = "${absPath}:/migrations"

Write-Host "Marking auth migrations as applied (version 3)..."
Write-Host "DB: postgres@$network"
Write-Host ""

docker run --rm `
    -v $volume `
    --network $network `
    migrate/migrate:v4.18.1 `
    -path=/migrations `
    "-database=$authDbUrl" `
    force 3

if ($LASTEXITCODE -ne 0) {
    throw "migrate force failed (exit code $LASTEXITCODE)"
}

Write-Host ""
Write-Host "Done. Run:"
Write-Host "  .\scripts\migrate.ps1"
