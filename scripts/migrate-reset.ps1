# Full DB schema reset (dev only - deletes ALL data in b2db)
param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"

if (-not $Force) {
    Write-Host "This will delete ALL tables and data in b2db (Docker container b2-postgres)."
    $answer = Read-Host "Continue? (yes/no)"
    if ($answer -ne "yes") {
        Write-Host "Cancelled."
        exit 0
    }
}

$container = "b2-postgres"
$running = docker ps --filter "name=$container" --format "{{.Names}}" 2>$null

if ($running -ne $container) {
    Write-Host "Container $container is not running. Start Postgres:"
    Write-Host "  cd services"
    Write-Host "  docker compose up -d postgres"
    exit 1
}

Write-Host "Resetting public schema in $container..."
docker exec $container psql -U b2user -d b2db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO b2user; GRANT ALL ON SCHEMA public TO public;"

if ($LASTEXITCODE -ne 0) {
    throw "Failed to reset database"
}

Write-Host "Done. Run migrations (uses same Docker DB):"
Write-Host "  .\scripts\migrate.ps1"
