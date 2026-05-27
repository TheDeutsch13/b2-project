param(
    [Parameter(Mandatory = $true)]
    [string]$Email
)

$ErrorActionPreference = "Stop"

$dbUrl = "postgres://b2user:b2password@localhost:5432/b2db?sslmode=disable"
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
    Write-Host "psql not found — copy the SQL above and run it manually."
}
