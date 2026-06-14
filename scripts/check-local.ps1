$ErrorActionPreference = "Stop"

Set-Location (Join-Path $PSScriptRoot "..")

Write-Host "==> gofmt check"
$goFiles = Get-ChildItem -Recurse -Filter *.go -File |
    Where-Object { $_.FullName -notmatch "\\.git\\" } |
    ForEach-Object { $_.FullName }

if ($goFiles.Count -gt 0) {
    $unformatted = & gofmt -l @goFiles
    if ($unformatted) {
        Write-Host "The following Go files need gofmt:"
        $unformatted | ForEach-Object { Write-Host $_ }
        exit 1
    }
}

Write-Host "==> go test ./..."
go test ./...

Write-Host "==> go vet ./..."
go vet ./...

Write-Host "==> docker compose config"
docker compose config | Out-Null

Write-Host "All local checks passed."
