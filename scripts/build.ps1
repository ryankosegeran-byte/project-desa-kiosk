# scripts/build.ps1
# Script to compile Go binaries and build frontend assets for deployment.

# Ensure output directory exists
$buildDir = "build"
if (!(Test-Path $buildDir)) {
    New-Item -ItemType Directory -Path $buildDir | Out-Null
}

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  MULAI PROSES KOMPILASI & BUILD KIOSK DESA   " -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan

# 1. Build Kiosk UI (React + Vite)
Write-Host "`n[1/4] Membangun Kiosk UI (web/kiosk-ui)..." -ForegroundColor Yellow
Push-Location "web/kiosk-ui"
npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Error "Gagal membangun Kiosk UI!"
    Pop-Location
    exit 1
}
Pop-Location

# 2. Build Admin Dashboard (Astro + React)
Write-Host "`n[2/4] Membangun Admin Dashboard (web/dashboard)..." -ForegroundColor Yellow
Push-Location "web/dashboard"
npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Error "Gagal membangun Admin Dashboard!"
    Pop-Location
    exit 1
}
Pop-Location

# 3. Compile Local Kiosk Go Binary (Windows Target)
Write-Host "`n[3/4] Mengompilasi Binary Go Kiosk (Windows)..." -ForegroundColor Yellow
$Env:GOOS = "windows"
$Env:GOARCH = "amd64"
go build -ldflags="-s -w" -o "$buildDir/kiosk.exe" ./kiosk/cmd/kiosk
if ($LASTEXITCODE -ne 0) {
    Write-Error "Gagal mengompilasi binary kiosk!"
    exit 1
}
Write-Host "Hasil: $buildDir/kiosk.exe" -ForegroundColor Green

# 4. Compile Hub Server Go Binary (Linux Target for serv00)
Write-Host "`n[4/4] Mengompilasi Binary Go Server Hub (Linux)..." -ForegroundColor Yellow
$Env:GOOS = "linux"
$Env:GOARCH = "amd64"
go build -ldflags="-s -w" -o "$buildDir/server" ./server/cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Error "Gagal mengompilasi binary server!"
    exit 1
}
Write-Host "Hasil: $buildDir/server" -ForegroundColor Green

# Restore defaults
$Env:GOOS = ""
$Env:GOARCH = ""

Write-Host "`n===============================================" -ForegroundColor Green
Write-Host "  SEMUA PROSES BUILD BERHASIL DISELESAIKAN!    " -ForegroundColor Green
Write-Host "  Semua file output disimpan di folder: /$buildDir" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Green
