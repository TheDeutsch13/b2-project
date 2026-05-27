param(
    [ValidateSet("up", "down")]
    [string]$Direction = "up"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$container = "b2-postgres"

$network = docker inspect -f '{{range $k, $v := .NetworkSettings.Networks}}{{$k}}{{end}}' $container 2>$null
if (-not $network) {
    Write-Host "Container $container is not running. Start Postgres:"
    Write-Host "  cd services"
    Write-Host "  docker compose up -d postgres"
    exit 1
}

# Same DB as migrate-reset (Docker postgres service, NOT host.docker.internal)
$authDbUrl = "postgres://b2user:b2password@postgres:5432/b2db?sslmode=disable&x-migrations-table=auth_schema_migrations"
$productDbUrl = "postgres://b2user:b2password@postgres:5432/b2db?sslmode=disable&x-migrations-table=product_schema_migrations"

function Invoke-Migrate {
    param(
        [string]$Path,
        [string]$DatabaseUrl,
        [string[]]$ExtraArgs
    )

    $absPath = (Resolve-Path $Path).Path
    $volume = "${absPath}:/migrations"

    $args = @(
        "run", "--rm",
        "-v", $volume,
        "--network", $network,
        "migrate/migrate:v4.18.1",
        "-path=/migrations",
        "-database=$DatabaseUrl"
    ) + $ExtraArgs

    & docker @args
    return $LASTEXITCODE
}

function Show-Version {
    param([string]$Label, [string]$Path, [string]$DatabaseUrl)

    Write-Host "Version ($Label):"
    $code = Invoke-Migrate -Path $Path -DatabaseUrl $DatabaseUrl -ExtraArgs @("version")
    if ($code -ne 0) {
        Write-Host "  (no version yet / first run)"
    }
}

function Run-Migrate {
    param([string]$Label, [string]$Path, [string]$DatabaseUrl)

    Write-Host "Running migrate $Direction for $Label..."
    $code = Invoke-Migrate -Path $Path -DatabaseUrl $DatabaseUrl -ExtraArgs @($Direction)

    if ($code -ne 0) {
        throw "Migration failed for $Label (exit code $code)"
    }

    Write-Host "  OK: $Label"
}

Write-Host "DB: postgres@$network (container $container)"
Write-Host ""

$authPath = Join-Path $root "migrations\auth"
$productPath = Join-Path $root "migrations\product"

Show-Version -Label "auth" -Path $authPath -DatabaseUrl $authDbUrl
Show-Version -Label "product" -Path $productPath -DatabaseUrl $productDbUrl
Write-Host ""

Run-Migrate -Label "auth" -Path $authPath -DatabaseUrl $authDbUrl
Run-Migrate -Label "product" -Path $productPath -DatabaseUrl $productDbUrl

Write-Host ""
Write-Host "Migrations completed successfully."
