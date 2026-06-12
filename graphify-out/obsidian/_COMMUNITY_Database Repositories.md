---
type: community
cohesion: 0.14
members: 14
---

# Database Repositories

**Cohesion:** 0.14 - loosely connected
**Members:** 14 nodes

## Members
- [[Package db]] - code - kiosk/db
- [[config_repo (db)]] - code - kiosk/db/config_repo.go
- [[db_test (db)]] - code - kiosk/db/db_test.go
- [[desa_repo (db)]] - code - server/db/desa_repo.go
- [[jenis_surat_repo (db)]] - code - kiosk/db/jenis_surat_repo.go
- [[nomor_surat_repo (db)]] - code - kiosk/db/nomor_surat_repo.go
- [[postgres (db)]] - code - server/db/postgres.go
- [[seeder (db)]] - code - kiosk/db/seeder.go
- [[sqlite (db)]] - code - kiosk/db/sqlite.go
- [[surat_repo (db)]] - code - kiosk/db/surat_repo.go
- [[sync_repo (db)]] - code - kiosk/db/sync_repo.go
- [[template_repo (db)]] - code - server/db/template_repo.go
- [[user_repo (db)]] - code - server/db/user_repo.go
- [[warga_repo (db)]] - code - kiosk/db/warga_repo.go

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/Database_Repositories
SORT file.name ASC
```

## Connections to other communities
- 1 edge to [[_COMMUNITY_Skill & Agent Configuration]]

## Top bridge nodes
- [[Package db]] - degree 14, connects to 1 community