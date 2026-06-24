# Pembaca RFID Fisik (ACR122U / PC/SC)

Kiosk kini mendukung pembaca RFID berbasis **PC/SC** seperti **ACR122U**
(mendukung kartu ISO14443 Type A & B). UID kartu yang terbaca dipublikasikan ke
pipeline SSE yang sama dengan mock-scan dan keyboard-wedge, sehingga UI kiosk
(`useRFID`) langsung menerimanya tanpa perubahan front-end.

## Cara Kerja

1. Backend Go memanggil `winscard.dll` Windows (tanpa CGO) untuk berkomunikasi
   dengan reader PC/SC.
2. Saat kartu ditempel, backend mengirim APDU standar `FF CA 00 00 00`
   (Get Data / UID).
3. UID diformat lalu di-`broker.Publish(uid)` -> dialirkan via
   `GET /api/rfid/events` (SSE) -> ditangkap `web/kiosk-ui/src/hooks/useRFID.ts`.
4. UI memanggil `GET /api/warga/rfid/{uid}` untuk mencari warga.

Implementasi: `kiosk/rfid/pcsc_windows.go` (Windows), `kiosk/rfid/pcsc_other.go`
(stub no-op platform lain), orchestrasi `kiosk/rfid/reader.go`.

## Prasyarat (Windows)

1. Colok ACR122U. Windows umumnya memasang driver CCID otomatis. Jika tidak,
   pasang driver resmi ACS.
2. Pastikan layanan **Smart Card** (`SCardSvr`) berjalan:
   ```powershell
   Get-Service SCardSvr
   Start-Service SCardSvr   # bila Stopped
   ```
3. Verifikasi reader terbaca OS:
   ```powershell
   certutil -scinfo   # akan melisting reader PC/SC yang terpasang
   ```

## Konfigurasi (.env)

```
KIOSK_RFID_PCSC_ENABLED=true   # aktif/nonaktif pembaca fisik
KIOSK_RFID_READER=             # opsional, filter substring nama reader (mis. "ACR122")
KIOSK_RFID_UID_FORMAT=hex      # hex | hex-colon | decimal
```

- `hex` -> contoh `04A1B2C3` (huruf besar, tanpa pemisah) — **default**.
- `hex-colon` -> contoh `04:A1:B2:C3`.
- `decimal` -> interpretasi little-endian -> contoh `3283263748`.

> Pastikan format ini cocok dengan nilai kolom `warga.rfid_uid`. Lookup di
> backend sudah case-insensitive (`LOWER(rfid_uid) = LOWER(?)`).

## Catatan e-KTP

UID chip e-KTP Indonesia umumnya **acak/anti-collision** (bukan NIK) dan dapat
berubah setiap penempelan untuk sebagian kartu. Gunakan UID sebagai pengidentifikasi
kartu yang dipetakan ke warga (kolom `rfid_uid`), bukan sebagai NIK. Bila UID
kartu tidak stabil, jalur **Masukkan NIK Manual** tetap tersedia.

## Uji Cepat

- Tanpa hardware (simulasi): `./scripts/mock-rfid.ps1 -Uid "04A1B2C3"`.
- Dengan ACR122U: jalankan `kiosk.exe`, tempel kartu; cek log
  `Kartu RFID terbaca dari pembaca fisik (PC/SC)` lalu UID muncul di UI.

## Troubleshooting

- Log `winscard.dll tidak tersedia` -> aktifkan layanan Smart Card.
- Log `tidak ada reader tersedia` -> reader belum terdeteksi OS / driver kurang.
- UID terbaca tapi warga tak ketemu -> sesuaikan `KIOSK_RFID_UID_FORMAT` atau
  perbarui `warga.rfid_uid` agar cocok.
- Nonaktifkan sementara: `KIOSK_RFID_PCSC_ENABLED=false` (keyboard-wedge & mock
  tetap jalan).