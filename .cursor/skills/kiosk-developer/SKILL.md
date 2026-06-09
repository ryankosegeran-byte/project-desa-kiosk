---
name: kiosk-developer
description: Developer playbook for the Desa Kiosk system. Helps run, test, build the backend services, and simulate the RFID reader.
category: development
tags:
  - golang
  - sqlite
  - kiosk
  - rfid
  - react
  - typescript
risk: safe
---

# 🌌 Kiosk Desa Developer Skill Playbook

Skill ini digunakan untuk mempermudah developer dan asisten AI dalam mengelola, mengembangkan, menguji, dan mensimulasikan sistem **Kiosk Pelayanan Administrasi Desa** yang bertipe *offline-first distributed system*.

---

## 🏗️ Gambaran Arsitektur

Sistem ini memiliki arsitektur monorepo dengan tiga komponen utama:
1.  **Local Kiosk Backend (`kiosk/`)**:
    *   **Teknologi**: Go + Chi Router + SQLite (local DB) + SSE (Server-Sent Events) untuk RFID Broker.
    *   **Peran**: Berjalan di hardware kiosk desa secara offline-first. Memproses scan kartu RFID (KTP), menyajikan assets frontend static, dan mencetak dokumen secara lokal.
2.  **Kiosk UI Frontend (`web/kiosk-ui/`)**:
    *   **Teknologi**: React + TypeScript + Vite + Lucide Icons.
    *   **Peran**: Tampilan layar interaktif kiosk desa dengan layout ramah sentuhan tangan (finger-friendly), dilengkapi virtual numeric keypad untuk input NIK warga.
3.  **Online Hub Backend (`server/`)**:
    *   **Teknologi**: Go + Chi Router + PostgreSQL + JWT Auth.
    *   **Peran**: Berjalan di cloud. Mengsinkronisasikan data warga, jenis surat, dan melayani dashboard administrasi.

---

## 🚀 Perintah Development Utama

Perintah ini harus dijalankan sesuai dengan lokasinya di dalam project monorepo.

### 1. Kiosk Backend & Database
Jalankan dari **root directory workspace** (`d:/PROJECT/MYPROJECT/project-desa-kiosk`):
*   Pastikan file `.env` diisi dengan `KIOSK_DESA_ID=desa-test-uuid`.
*   **Run local kiosk backend**:
    ```powershell
    go run ./kiosk/cmd/kiosk
    ```
    *Server berjalan di http://localhost:8080 dan otomatis membuat `data/kiosk.db` serta mengisi data seeder jika kosong.*
*   **Run unit tests**:
    ```powershell
    go test ./kiosk/api/... ./kiosk/db/...
    ```

### 2. Kiosk UI Frontend
Jalankan dari directory **`web/kiosk-ui/`**:
*   **Instalasi dependency**:
    ```powershell
    cd web/kiosk-ui
    npm install
    ```
*   **Run dev server** (dengan hot reload):
    ```powershell
    npm run dev
    ```
    *Dev server berjalan di http://localhost:5173 dan otomatis meneruskan API call ke backend local port 8080.*
*   **Build untuk produksi**:
    ```powershell
    npm run build
    ```
    *Kompilasi TypeScript dan Vite. File output secara otomatis disimpan ke `web/dist` (diatur di `vite.config.ts`) agar langsung dapat disajikan oleh backend Go.*

### 3. Online Server Backend
Jalankan dari **root directory workspace**:
*   **Run online server**:
    ```powershell
    go run ./server/cmd/server
    ```

---

## 🔄 Protokol Sinkronisasi (Sync Protocol)

Kiosk berjalan offline-first dan mencatat perubahan lokal di tabel database `sync_queue`. Developer harus memastikan background worker berjalan di Kiosk backend untuk melakukan sinkronisasi secara berkala (`KIOSK_SYNC_INTERVAL` detik):

1.  **Push Data (Kiosk -> Server):**
    *   Ambil data dari `sync_queue` yang belum diproses.
    *   Kirimkan payload `SyncPushPayload` ke endpoint server `POST /api/sync/push`.
    *   Jika sukses, tandai antrian lokal sebagai selesai menggunakan `MarkSynced`.
2.  **Pull Data (Server -> Kiosk):**
    *   Kiosk melakukan request `GET /api/sync/pull/warga` dan `GET /api/sync/pull/config` untuk mengunduh pembaruan data warga dan template surat terbaru dari pusat.

---

## 💳 Simulasi Scan RFID (Tap KTP)

Kiosk Backend memiliki Broker RFID yang memancarkan event tap kartu secara real-time via Server-Sent Events (SSE) di `/api/rfid/events`. 

Untuk mensimulasikan tap kartu RFID warga tanpa hardware asli:
1.  Gunakan script pembantu PowerShell yang tersedia:
    ```powershell
    ./scripts/mock-rfid.ps1 -Uid "1234567890"
    ```

### Warga Ter-Seeding (Untuk Pengujian)
Saat database SQLite lokal kosong, sistem akan otomatis melakukan *seeding* data warga berikut:
*   **Budi Santoso** (RFID UID: `1234567890`, NIK: `3201234567890001`)
*   **Siti Aminah** (RFID UID: `0987654321`, NIK: `3201234567890002`)

---

## 🛠️ Panduan Pemecahan Masalah (Troubleshooting)

### 1. Halaman Kiosk UI 404 Tidak Ditemukan
Jika Anda membuka `http://localhost:8080/` dan mendapatkan halaman kosong/404, pastikan Anda telah menjalankan build frontend React terlebih dahulu (`cd web/kiosk-ui` -> `npm run build`). Ini akan menghasilkan aset statis di `web/dist` yang dibaca oleh server Go.

### 2. Port 8080 atau 5173 Terkunci (Address Already In Use)
Jika Anda mendapatkan error port sedang digunakan, matikan proses zombie di background:
*   Untuk Go: Cari proses `kiosk.exe` atau jalankan `Stop-Process -Name kiosk` di PowerShell.
*   Untuk Node/Vite: Hentikan server vite yang lama dengan `Ctrl + C` pada terminal Anda.
