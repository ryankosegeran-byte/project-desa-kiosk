# Dashboard Admin — Desa Kiosk

Panel admin (SPA) untuk mengelola data warga, template surat, penomoran surat,
dan memantau kiosk. Dibangun dengan **React 19 + Vite + TypeScript** (React Router).

## Menjalankan

Semua perintah dijalankan dari folder `web/dashboard`:

| Perintah | Aksi |
| :--- | :--- |
| `npm install` | Pasang dependensi |
| `npm run dev` | Dev server di `localhost:4321` (proxy `/api` → Go server `:3000`) |
| `npm run build` | Build produksi ke `./dist/` |
| `npm run preview` | Pratinjau hasil build secara lokal |

## Struktur

```text
src/
├── components/   # Komponen halaman & modal (Templates, Warga, FormVariabel, NomorSurat, …)
├── layouts/      # Kerangka dashboard
├── lib/          # Helper API (auth, request)
├── styles/       # CSS global
├── App.tsx       # Definisi route
└── main.tsx      # Entry point
```

## Catatan

- Backend Go melayani build statis (`dist/`) di serv00, sehingga `/api` same-origin di produksi.
- Preview surat di dashboard memakai **PDF pendamping** yang diunggah admin (serv00 tidak merender DOCX);
  surat asli dicetak di kiosk via Microsoft Word.
