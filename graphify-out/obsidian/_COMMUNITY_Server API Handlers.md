---
type: community
cohesion: 0.11
members: 18
---

# Server API Handlers

**Cohesion:** 0.11 - loosely connected
**Members:** 18 nodes

## Members
- [[Package api]] - code - kiosk/api
- [[api_test (api)]] - code - kiosk/api/api_test.go
- [[auth_handler (api)]] - code - server/api/auth_handler.go
- [[dashboard_handler (api)]] - code - server/api/dashboard_handler.go
- [[desa_handler (api)]] - code - server/api/desa_handler.go
- [[helpers (api)]] - code - server/api/helpers.go
- [[jenis_surat_handler (api)]] - code - server/api/jenis_surat_handler.go
- [[nomor_surat_handler (api)]] - code - server/api/nomor_surat_handler.go
- [[ocr_handler (api)]] - code - server/api/ocr_handler.go
- [[ocr_status_handler (api)]] - code - server/api/ocr_status_handler.go
- [[router (api)]] - code - kiosk/api/router.go
- [[surat (api)]] - code - kiosk/api/surat.go
- [[surat_handler (api)]] - code - server/api/surat_handler.go
- [[sync_handler (api)]] - code - server/api/sync_handler.go
- [[template_handler (api)]] - code - server/api/template_handler.go
- [[user_handler (api)]] - code - server/api/user_handler.go
- [[warga (api)]] - code - kiosk/api/warga.go
- [[warga_handler (api)]] - code - server/api/warga_handler.go

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/Server_API_Handlers
SORT file.name ASC
```

## Connections to other communities
- 1 edge to [[_COMMUNITY_Skill & Agent Configuration]]

## Top bridge nodes
- [[Package api]] - degree 18, connects to 1 community