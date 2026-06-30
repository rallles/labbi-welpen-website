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

if ($env:SKIP_ASSET_CHECK -eq "1") {
    Write-Host "==> Asset-Check uebersprungen (SKIP_ASSET_CHECK=1)"
} else {
    Write-Host "==> asset check"
    # Suche /static/... Referenzen in internal/templates und static/css und
    # pruefe, ob die referenzierten Dateien existieren. Es wird nichts geloescht.
    $searchDirs = @("internal/templates", "static/css") | Where-Object { Test-Path $_ }
    $missing = @()
    if ($searchDirs.Count -gt 0) {
        $refs = Select-String -Path (Get-ChildItem -Recurse -File $searchDirs).FullName -Pattern '/static/[A-Za-z0-9_./-]+' -AllMatches |
            ForEach-Object { $_.Matches } |
            ForEach-Object { $_.Value } |
            Sort-Object -Unique
        foreach ($ref in $refs) {
            $file = "static/" + ($ref -replace '^/static/', '')
            if (-not (Test-Path -LiteralPath $file)) {
                $missing += "$ref  ->  $file"
            }
        }
    }
    if ($missing.Count -gt 0) {
        Write-Host "==> Fehlende Assets ($($missing.Count)):"
        $missing | ForEach-Object { Write-Host "  $_" }
        Write-Host "Hinweis: ZIP-Arbeitsstaende sind manchmal absichtlich ohne Bilder."
        Write-Host "Mit SKIP_ASSET_CHECK=1 kann der Check uebersprungen werden."
        exit 1
    }
    Write-Host "    Alle referenzierten Assets sind vorhanden."
}

Write-Host "All local checks passed."
