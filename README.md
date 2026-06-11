# Sistem Kiosk Desa

Offline-first distributed system untuk pelayanan administrasi desa. Warga datang ke kiosk, tap KTP pada RFID reader, pilih jenis surat, dan surat langsung dicetak.

## Arsitektur

```
Cloud (Online)                      Kiosk (Per Desa)
┌──────────────────┐                ┌──────────────────┐
│ Cloudflare Pages │                │ Go Local API     │
│ (Dashboard)      │                │ SQLite           │
│                  │◄─── sync ────►│ RFID Reader      │
│ Go + Chi Backend │                │ Chrome Kiosk     │
│ PostgreSQL       │                │ Printer A4       │
└──────────────────┘                └──────────────────┘
```

## Tech Stack

| Komponen | Teknologi |
|----------|-----------|
| Kiosk Backend | Go + Chi + SQLite |
| Online Backend | Go + Chi + PostgreSQL |
| Kiosk UI | React + Vite |
| Dashboard | Astro + React Islands |
| PDF | chromedp (headless Chrome) |
| Auth | JWT |
| Hosting | Cloudflare Pages + serv00 |

## Struktur Monorepo

```
project-desa-kiosk/
├── go.work              # Go workspace
├── internal/            # Shared models & auth
├── kiosk/               # Local kiosk backend
├── server/              # Online backend
├── web/
│   ├── kiosk-ui/        # React kiosk app
│   └── dashboard/       # Astro admin dashboard
├── cmd/
│   ├── kiosk/           # Kiosk entry point
│   └── server/          # Server entry point
├── scripts/             # Setup & build scripts
└── deployments/         # Deploy configs
```

## Quick Start

### Kiosk (Development)

```bash
# Set environment
cp .env.example .env
# Edit .env with your desa_id

# Run kiosk backend
go run ./cmd/kiosk

# Buka browser ke http://localhost:8080
```

### Server (Development)

```bash
# Set environment
export DATABASE_URL="postgres://..."
export JWT_SECRET="your-secret"

# Run server
go run ./cmd/server
```

## Lisensi

Private - All rights reserved.
