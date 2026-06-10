# Catatan Konfigurasi Kiosk

## Cara Mendapatkan KIOSK_DESA_ID dan KIOSK_API_KEY

### 1. Dapatkan KIOSK_DESA_ID

- Buka **Dashboard** di `http://localhost:4321`
- Masuk ke menu **Kelola Desa** (sidebar Admin Panel)
- Buat desa baru atau pilih desa yang sudah ada
- Salin **UUID desa** yang ditampilkan — ini adalah nilai `KIOSK_DESA_ID`

### 2. Dapatkan KIOSK_API_KEY

- Buka halaman **Status Kiosk** di `http://localhost:4321/kiosk-status`
- Klik tombol **"+ Daftarkan Kiosk Baru"** (kanan atas)
- Isi nama kiosk dan pilih desa yang sesuai
- Setelah berhasil didaftarkan, sistem akan men-generate **API Key** unik
- Salin API key tersebut — ini adalah nilai `KIOSK_API_KEY`

### 3. Set di File `.env`

Buka file `.env` di root project, lalu isi:

```env
KIOSK_DESA_ID=<uuid-desa-dari-kelola-desa>
KIOSK_API_KEY=<api-key-dari-pendaftaran-kiosk>
KIOSK_SERVER_URL=http://localhost:4321
```

### 4. Restart Kiosk

Setelah mengubah `.env`, restart kiosk backend agar konfigurasi terbaca:

```powershell
.\kiosk.exe
```

---

## Catatan Penting

- **KIOSK_DESA_ID** harus UUID valid yang terdaftar di database PostgreSQL (bukan placeholder seperti `desa-test-uuid`)
- **KIOSK_API_KEY** harus API key yang di-issue oleh dashboard setelah registrasi unit kiosk
- **KIOSK_SERVER_URL** harus diisi agar sync engine aktif
- Tanpa ketiga nilai ini, sinkronisasi kiosk ↔ server **tidak akan berjalan**

## Alur Singkat

```
Kelola Desa → buat desa → dapat UUID
     ↓
Status Kiosk → daftarkan kiosk → dapat API Key
     ↓
.env → isi KIOSK_DESA_ID + KIOSK_API_KEY + KIOSK_SERVER_URL
     ↓
Restart kiosk → sync aktif ✓
```
