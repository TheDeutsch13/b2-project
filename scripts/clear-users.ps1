param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"

if (-not $Force) {
    Write-Host "This will DELETE ONLY user-related data in b2db (keeps products/categories)."
    Write-Host "Will remove: users, refresh_tokens, orders(+items), support threads/messages, and product embedded reviews."
    Write-Host ""
    Write-Host "NOTE: Go services on Windows connect to localhost:5432 (host), NOT docker exec."
    $answer = Read-Host "Continue? (yes/no)"
    if ($answer -ne "yes") {
        Write-Host "Cancelled."
        exit 0
    }
}

# Same DB as auth-service / product-service on Windows (localhost:5432)
$databaseUrl = "postgres://b2user:b2password@host.docker.internal:5432/b2db?sslmode=disable"

Write-Host "Clearing user-related data (database used by Go on localhost:5432)..."

$sql = @"
BEGIN;

TRUNCATE TABLE order_items RESTART IDENTITY CASCADE;
TRUNCATE TABLE orders RESTART IDENTITY CASCADE;

TRUNCATE TABLE support_messages RESTART IDENTITY CASCADE;
TRUNCATE TABLE support_threads RESTART IDENTITY CASCADE;

UPDATE products
SET reviews = '[]'::jsonb,
    rating_avg = 0,
    rating_count = 0
WHERE reviews IS NOT NULL AND reviews <> '[]'::jsonb;

TRUNCATE TABLE refresh_tokens RESTART IDENTITY CASCADE;
TRUNCATE TABLE users RESTART IDENTITY CASCADE;

COMMIT;
"@

docker run --rm postgres:16 psql $databaseUrl -v ON_ERROR_STOP=1 -c $sql

if ($LASTEXITCODE -ne 0) {
    throw "Failed to clear users"
}

Write-Host ""
Write-Host "Verification:"
docker run --rm postgres:16 psql $databaseUrl -c "SELECT count(*) AS users FROM users; SELECT count(*) AS support_messages FROM support_messages; SELECT count(*) AS products FROM products;"

Write-Host ""
Write-Host "Done. Log in again should fail until you register a new account."
Write-Host "If you need admin: .\scripts\grant-admin.ps1 -Email your@email.com"
