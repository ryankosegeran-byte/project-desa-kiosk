---
type: community
cohesion: 0.29
members: 7
---

# Data Models

**Cohesion:** 0.29 - loosely connected
**Members:** 7 nodes

## Members
- [[Package models]] - code - internal/models
- [[desa (models)]] - code - internal/models/desa.go
- [[jenis_surat (models)]] - code - internal/models/jenis_surat.go
- [[surat (models)]] - code - internal/models/surat.go
- [[sync (models)]] - code - internal/models/sync.go
- [[user (models)]] - code - internal/models/user.go
- [[warga (models)]] - code - internal/models/warga.go

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/Data_Models
SORT file.name ASC
```
