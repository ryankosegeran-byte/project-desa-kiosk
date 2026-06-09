# scripts/setup-kiosk.ps1
# Windows Local Kiosk Auto-Installer & Setup Script.
# Run this on the target kiosk machine to prepare folders, database, env, and auto-start.

Write-Host "==============================================" -ForegroundColor Cyan
Write-Host "   PANDUAN SETUP AUTOMATIS KIOSK DESA LOKAL   " -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan

# 1. Create directory structure
$baseDir = "C:\KioskDesa"
$dataDir = "$baseDir\data"
$pdfDir = "$baseDir\data\pdf"
$binDir = "$baseDir\bin"

Write-Host "`n[1/5] Membuat direktori aplikasi di $baseDir..." -ForegroundColor Yellow
if (!(Test-Path $baseDir)) { New-Item -ItemType Directory -Path $baseDir | Out-Null }
if (!(Test-Path $dataDir)) { New-Item -ItemType Directory -Path $dataDir | Out-Null }
if (!(Test-Path $pdfDir)) { New-Item -ItemType Directory -Path $pdfDir | Out-Null }
if (!(Test-Path $binDir)) { New-Item -ItemType Directory -Path $binDir | Out-Null }
Write-Host "Direktori berhasil dibuat." -ForegroundColor Green

# 2. Copy kiosk.exe to C:\KioskDesa\bin
Write-Host "`n[2/5] Menyalin file executables..." -ForegroundColor Yellow
if (Test-Path "build\kiosk.exe") {
    Copy-Item "build\kiosk.exe" "$binDir\kiosk.exe" -Force
    Write-Host "kiosk.exe disalin ke $binDir." -ForegroundColor Green
} else {
    Write-Host "PERINGATAN: build\kiosk.exe tidak ditemukan. Pastikan Anda telah menjalankan .\scripts\build.ps1 terlebih dahulu." -ForegroundColor Red
}

# 3. Setup Environment File (.env)
Write-Host "`n[3/5] Menyiapkan konfigurasi .env..." -ForegroundColor Yellow
$envPath = "$baseDir\.env"
if (Test-Path $envPath) {
    Write-Host "File .env sudah ada, melewati pembuatan baru." -ForegroundColor Green
} else {
    $serverUrl = Read-Host "Masukkan URL Online Hub Server (default: http://localhost:3000)"
    if ([string]::IsNullOrEmpty($serverUrl)) { $serverUrl = "http://localhost:3000" }
    
    $apiKey = Read-Host "Masukkan API Key Kiosk Anda (didapat dari dashboard superadmin)"
    if ([string]::IsNullOrEmpty($apiKey)) { $apiKey = "kiosk_placeholder_api_key" }
    
    $kioskName = Read-Host "Masukkan Nama Kiosk (default: Kiosk Balai Desa)"
    if ([string]::IsNullOrEmpty($kioskName)) { $kioskName = "Kiosk Balai Desa" }

    $envContent = @"
PORT=8080
ENV=production
DB_PATH=$dataDir\kiosk.db
SERVER_URL=$serverUrl
API_KEY=$apiKey
KIOSK_NAME=$kioskName
KIOSK_PRINT_CMD=SumatraPDF.exe
PDF_TEMP_DIR=$pdfDir
"@
    $envContent | Out-File -FilePath $envPath -Encoding utf8
    Write-Host "File konfigurasi .env dibuat di $envPath." -ForegroundColor Green
}

# 4. Check & Download SumatraPDF (silent printing tool)
Write-Host "`n[4/5] Memeriksa instalasi SumatraPDF..." -ForegroundColor Yellow
$sumatraPath = "C:\Users\$env:USERNAME\AppData\Local\SumatraPDF\SumatraPDF.exe"
if (Test-Path $sumatraPath) {
    Write-Host "SumatraPDF ditemukan di: $sumatraPath" -ForegroundColor Green
} else {
    Write-Host "Mengunduh SumatraPDF Portable untuk silent printing..." -ForegroundColor Cyan
    $sumatraUrl = "https://www.sumatrapdfreader.org/dl/SumatraPDF-3.5.2-64.exe"
    $sumatraDest = "$binDir\SumatraPDF.exe"
    try {
        Invoke-WebRequest -Uri $sumatraUrl -OutFile $sumatraDest
        Write-Host "SumatraPDF berhasil diunduh dan dipasang di $sumatraDest." -ForegroundColor Green
    } catch {
        Write-Host "PERINGATAN: Gagal mengunduh SumatraPDF otomatis. Silakan unduh manual dan letakkan di $binDir\SumatraPDF.exe." -ForegroundColor Red
    }
}

# 5. Create Startup Shortcut
Write-Host "`n[5/5] Mengatur Autostart Kiosk saat Windows Boot..." -ForegroundColor Yellow
$startupFolder = [System.IO.Path]::Combine([Environment]::GetFolderPath('Startup'))
$shortcutPath = "$startupFolder\KioskDesa.lnk"

try {
    $WshShell = New-Object -ComObject WScript.Shell
    $Shortcut = $WshShell.CreateShortcut($shortcutPath)
    $Shortcut.TargetPath = "$binDir\kiosk.exe"
    $Shortcut.WorkingDirectory = $baseDir
    $Shortcut.Save()
    Write-Host "Startup shortcut berhasil dibuat di folder Startup Windows." -ForegroundColor Green
} catch {
    Write-Host "Gagal mengatur startup otomatis. Silakan buat shortcut secara manual ke folder Startup." -ForegroundColor Red
}

Write-Host "`n==============================================" -ForegroundColor Green
Write-Host "     SETUP KIOSK LOKAL SELESAI & SUKSES!      " -ForegroundColor Green
Write-Host "  Silakan jalankan kiosk via shortcut atau run:  " -ForegroundColor Green
Write-Host "  & '$binDir\kiosk.exe' dari folder $baseDir   " -ForegroundColor Green
Write-Host "==============================================" -ForegroundColor Green
