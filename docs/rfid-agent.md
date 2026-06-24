# RFID Agent (Jembatan ACR122U untuk Admin Panel Online)

Admin panel disajikan online, sedangkan reader ACR122U adalah perangkat USB
(PC/SC) di **laptop PIC**. Browser tidak bisa mengakses reader USB secara
langsung, jadi diperlukan agen lokal kecil: `rfid-agent.exe`.

```
ACR122U (USB) -> rfid-agent.exe (localhost:8088, SSE) -> browser admin panel
```

Agen ini memakai package RFID yang sama dengan kiosk (tanpa database / desa_id),
sehingga ringan dan cukup dijalankan di laptop PIC.

## Menjalankan (laptop PIC)

1. Pastikan layanan Smart Card aktif:
   ```powershell
   Get-Service SCardSvr
   Start-Service SCardSvr   # bila Stopped
   ```
2. Colok ACR122U.
3. Jalankan agen:
   ```powershell
   .\rfid-agent.exe
   ```
   Lalu buka Registrasi Warga -> langkah 3. Badge "Pembaca ACR122U (agen lokal)
   terhubung" akan menyala saat agen aktif. Tempel KTP -> UID otomatis terisi.

Agen tetap opsional: bila tidak dijalankan, mode **plug-n-play (keyboard wedge)**
tetap berfungsi seperti sebelumnya.

## Konfigurasi (opsional, environment variable)

| Variabel | Default | Keterangan |
|---|---|---|
| `RFID_AGENT_LISTEN` | `:8088` | Alamat listen HTTP lokal |
| `RFID_AGENT_UID_FORMAT` | `hex` | `hex` \| `hex-colon` \| `decimal` |
| `RFID_AGENT_READER` | (kosong) | Filter substring nama reader, mis. `ACR122` |

Di sisi admin panel, arahkan ke agen via `VITE_RFID_BRIDGE_URL`
(default `http://localhost:8088`).

## Endpoint Agen

- `GET  /api/rfid/events` — SSE stream UID (dipakai admin panel).
- `POST /api/rfid/mock`  — `{"uid":"04A1B2C3"}` untuk uji tanpa hardware.
- `GET  /healthz`        — status agen.

## Uji Cepat

```powershell
# tanpa hardware:
Invoke-RestMethod -Uri http://localhost:8088/api/rfid/mock -Method Post `
  -Body (@{uid="04A1B2C3"} | ConvertTo-Json) -ContentType "application/json"
```
UID akan langsung muncul di langkah 3 admin panel yang sedang terbuka.

## Catatan Format UID

Pastikan format UID agen sama dengan yang tersimpan di data warga (`rfid_uid`).
Default `hex` (huruf besar, tanpa pemisah), mis. `04A1B2C3`. Lihat juga
`docs/rfid-acr122u.md`.