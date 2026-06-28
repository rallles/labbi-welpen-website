$ErrorActionPreference = "Stop"

Set-Location (Join-Path $PSScriptRoot "..")

Write-Host "==> gofmt check"
$goFiles = Get-ChildItem -Recurse -Filter *.go -File |
    Where-Object { $_.FullName -notmatch "\\.git\\" } |
    ForEach-Object { $_.FullName }

if ($goFiles.Count -gt 0) {
    $unformatted = & gofmt -l @goFiles
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
    if ($unformatted) {
        Write-Host "The following Go files need gofmt:"
        $unformatted | ForEach-Object { Write-Host $_ }
        exit 1
    }
}

Write-Host "==> go test ./..."
go test ./...
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "==> go vet ./..."
go vet ./...
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "==> docker compose config (validation only; no containers are started)"
docker compose config | Out-Null
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "All local checks passed."
