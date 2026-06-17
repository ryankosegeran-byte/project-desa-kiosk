---
type: community
cohesion: 0.11
members: 32
---

# Server Router & Middleware

**Cohesion:** 0.11 - loosely connected
**Members:** 32 nodes

## Members
- [[.Handler()]] - code - kiosk/api/router.go
- [[.handleStatus()]] - code - kiosk/api/router.go
- [[Broker]] - code - kiosk/api/router.go
- [[Broker_2]] - code - kiosk/rfid/listener.go
- [[CORS()]] - code - kiosk/api/router.go
- [[CORS()_1]] - code - server/middleware/cors.go
- [[Config]] - code - kiosk/api/router.go
- [[ConfigRepository]] - code - kiosk/api/router.go
- [[HandleMockScan()]] - code - kiosk/rfid/listener.go
- [[Handler]] - code - kiosk/api/router.go
- [[Handler_3]] - code - server/middleware/cors.go
- [[HandlerFunc]] - code - kiosk/rfid/listener.go
- [[JenisSuratRepository]] - code - kiosk/api/router.go
- [[LoggerMiddleware()]] - code - kiosk/api/router.go
- [[MockScanRequest]] - code - kiosk/rfid/listener.go
- [[NewServer()]] - code - kiosk/api/router.go
- [[NomorSuratRepository]] - code - kiosk/api/router.go
- [[PDFGenerator]] - code - kiosk/api/router.go
- [[Printer]] - code - kiosk/api/router.go
- [[Request]] - code - kiosk/api/router.go
- [[ResponseWriter]] - code - kiosk/api/router.go
- [[Router]] - code - kiosk/api/router.go
- [[ServeEvents()]] - code - kiosk/rfid/listener.go
- [[Server_1]] - code - kiosk/api/router.go
- [[SuratRepository]] - code - kiosk/api/router.go
- [[WargaRepository]] - code - kiosk/api/router.go
- [[cors.go]] - code - server/middleware/cors.go
- [[listener.go]] - code - kiosk/rfid/listener.go
- [[router.go]] - code - kiosk/api/router.go
- [[sendError()]] - code - kiosk/api/router.go
- [[sendJSON()]] - code - kiosk/api/router.go
- [[setupStaticFileServer()]] - code - kiosk/api/router.go

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/Server_Router__Middleware
SORT file.name ASC
```

## Connections to other communities
- 4 edges to [[_COMMUNITY_Sync Handlers]]

## Top bridge nodes
- [[HandlerFunc]] - degree 10, connects to 1 community