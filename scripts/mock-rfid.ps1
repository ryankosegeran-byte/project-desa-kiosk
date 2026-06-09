# scripts/mock-rfid.ps1
# Script untuk mensimulasikan scan RFID warga ke local Kiosk API.
# Contoh penggunaan:
#   .\scripts\mock-rfid.ps1 -Uid "1234567890"

param (
    [string]$Uid = "1234567890"
)

$uri = "http://localhost:8080/api/rfid/mock"
$body = @{ uid = $Uid } | ConvertTo-Json

Write-Host "Mengirim mock scan RFID UID: $Uid ke $uri..." -ForegroundColor Cyan

try {
    # Send HTTP POST request
    $response = Invoke-RestMethod -Uri $uri -Method Post -Body $body -ContentType "application/json"
    
    if ($response.status -eq "success") {
        Write-Host "Respon Server: SUCCESS!" -ForegroundColor Green
        Write-Host "Kartu RFID warga dengan UID '$Uid' berhasil di-scan." -ForegroundColor Green
    } else {
        Write-Host "Gagal mengirim scan RFID. Respon Server: " -ForegroundColor Red -NoNewline
        Write-Host ($response | ConvertTo-Json -Compress) -ForegroundColor Yellow
    }
} catch {
    Write-Host "ERROR: Gagal terhubung ke server Kiosk Desa." -ForegroundColor Red
    Write-Host "Pastikan backend Kiosk sedang berjalan di port 8080 (coba jalankan 'go run ./kiosk/cmd/kiosk')." -ForegroundColor Yellow
    Write-Host "Pesan Error: $_" -ForegroundColor Red
}
