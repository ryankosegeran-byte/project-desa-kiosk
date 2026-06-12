---
type: community
cohesion: 0.33
members: 6
---

# Database Migrations

**Cohesion:** 0.33 - loosely connected
**Members:** 6 nodes

## Members
- [[001_init (migrations)]] - code - kiosk/db/migrations/001_init.sql
- [[002_add_warga_sync_columns (migrations)]] - code - kiosk/db/migrations/002_add_warga_sync_columns.sql
- [[002_warga_draft (migrations)]] - code - server/db/migrations/002_warga_draft.sql
- [[003_nomor_surat_batch (migrations)]] - code - kiosk/db/migrations/003_nomor_surat_batch.sql
- [[003_nomor_surat_config (migrations)]] - code - server/db/migrations/003_nomor_surat_config.sql
- [[Package migrations]] - code - kiosk/db/migrations

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/Database_Migrations
SORT file.name ASC
```
