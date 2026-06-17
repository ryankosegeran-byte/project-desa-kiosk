---
type: community
cohesion: 0.09
members: 40
---

# Sync Handlers

**Cohesion:** 0.09 - loosely connected
**Members:** 40 nodes

## Members
- [[.Handler()_1]] - code - server/api/router.go
- [[.handleListNomorSuratConfig()]] - code - server/api/nomor_surat_handler.go
- [[.handleSyncPullConfig()]] - code - server/api/sync_handler.go
- [[.handleSyncPullNomorSurat()]] - code - server/api/nomor_surat_handler.go
- [[.handleSyncPullWarga()]] - code - server/api/sync_handler.go
- [[.handleSyncPush()]] - code - server/api/sync_handler.go
- [[.handleUpdateNomorSuratConfig()]] - code - server/api/nomor_surat_handler.go
- [[AuthMiddleware()]] - code - server/middleware/auth.go
- [[Config_5]] - code - server/api/router.go
- [[Context_19]] - code - server/middleware/auth.go
- [[DB_9]] - code - server/api/router.go
- [[DesaRepository]] - code - server/api/router.go
- [[DesaRepository_2]] - code - server/middleware/auth.go
- [[GetKiosk()]] - code - server/middleware/auth.go
- [[Handler_1]] - code - server/api/router.go
- [[Handler_2]] - code - server/middleware/auth.go
- [[HasRole()]] - code - internal/auth/jwt.go
- [[JWTManager_1]] - code - server/api/router.go
- [[JWTManager_2]] - code - server/middleware/auth.go
- [[JenisSuratRepository_3]] - code - server/api/router.go
- [[Kiosk_2]] - code - server/middleware/auth.go
- [[KioskKeyMiddleware()]] - code - server/middleware/auth.go
- [[LoggerMiddleware()_1]] - code - server/api/router.go
- [[NewServer()_1]] - code - server/api/router.go
- [[Request_8]] - code - server/api/nomor_surat_handler.go
- [[Request_12]] - code - server/api/sync_handler.go
- [[ResponseWriter_8]] - code - server/api/nomor_surat_handler.go
- [[ResponseWriter_12]] - code - server/api/sync_handler.go
- [[RoleMiddleware()]] - code - server/middleware/auth.go
- [[Server_8]] - code - server/api/nomor_surat_handler.go
- [[Server_11]] - code - server/api/router.go
- [[Server_13]] - code - server/api/sync_handler.go
- [[Service]] - code - server/api/router.go
- [[SuratRepository_3]] - code - server/api/router.go
- [[TemplateRepository]] - code - server/api/router.go
- [[UserRepository]] - code - server/api/router.go
- [[WargaRepository_3]] - code - server/api/router.go
- [[auth.go]] - code - server/middleware/auth.go
- [[contextKey]] - code - server/middleware/auth.go
- [[router.go_1]] - code - server/api/router.go

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/Sync_Handlers
SORT file.name ASC
```

## Connections to other communities
- 5 edges to [[_COMMUNITY_API Helpers & Utilities]]
- 4 edges to [[_COMMUNITY_Server Router & Middleware]]
- 2 edges to [[_COMMUNITY_JWT Authentication]]

## Top bridge nodes
- [[auth.go]] - degree 6, connects to 1 community
- [[AuthMiddleware()]] - degree 6, connects to 1 community
- [[KioskKeyMiddleware()]] - degree 6, connects to 1 community
- [[RoleMiddleware()]] - degree 6, connects to 1 community
- [[.handleSyncPush()]] - degree 5, connects to 1 community