param(
    [Parameter(Mandatory = $true)]
    [string]$Email
)

$ErrorActionPreference = "Stop"

$hostDbUrl = "postgres://b2user:b2password@localhost:5432/b2db?sslmode=disable"
# From inside Docker container, "localhost" points to the container itself.
# Services running via `go run` use the host Postgres on localhost:5432, so we reach it via host.docker.internal.
$dockerDbUrl = "postgres://b2user:b2password@host.docker.internal:5432/b2db?sslmode=disable"
$escapedEmail = $Email.Replace("'", "''")
$sql = "UPDATE users SET role = 'admin' WHERE email = '$escapedEmail';"

Write-Host "Granting admin role to: $Email"
Write-Host ""
Write-Host "Run this SQL in PostgreSQL (pgAdmin or psql):"
Write-Host $sql
Write-Host ""
Write-Host "Important: log out and log in again so JWT gets the new role."
Write-Host ""

if (Get-Command psql -ErrorAction SilentlyContinue) {
    $env:PGPASSWORD = "b2password"
    psql -h localhost -U b2user -d b2db -c $sql
    Write-Host "Done. Re-login in the browser."
} else {
    Write-Host "psql not found - applying the UPDATE via docker postgres:16..."
    & docker run --rm postgres:16 psql "$dockerDbUrl" -v ON_ERROR_STOP=1 -c "$sql" | Out-Host
    Write-Host "Done. Re-login in the browser."
}
