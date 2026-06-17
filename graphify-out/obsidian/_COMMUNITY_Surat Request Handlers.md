---
type: community
cohesion: 0.15
members: 24
---

# Surat Request Handlers

**Cohesion:** 0.15 - loosely connected
**Members:** 24 nodes

## Members
- [[.GeneratePDF()]] - code - kiosk/print/pdf.go
- [[.handleCreateSurat()]] - code - kiosk/api/surat.go
- [[.handleGetJenisSuratSchema()]] - code - kiosk/api/surat.go
- [[.handleGetSurat()]] - code - kiosk/api/surat.go
- [[.handleGetTemplate()]] - code - kiosk/api/surat.go
- [[.handleListJenisSurat()]] - code - kiosk/api/surat.go
- [[.handleListTodaySurat()]] - code - kiosk/api/surat.go
- [[.handleNomorSuratStatus()]] - code - kiosk/api/surat.go
- [[.handlePrintSurat()]] - code - kiosk/api/surat.go
- [[Context_7]] - code - kiosk/print/pdf.go
- [[FormatIndonesianDate()]] - code - kiosk/print/pdf.go
- [[NewPDFGenerator()]] - code - kiosk/print/pdf.go
- [[PDFGenerator_1]] - code - kiosk/print/pdf.go
- [[Request_1]] - code - kiosk/api/surat.go
- [[ResponseWriter_1]] - code - kiosk/api/surat.go
- [[Server_2]] - code - kiosk/api/surat.go
- [[T_2]] - code - kiosk/print/print_test.go
- [[TestFormatIndonesianDate()]] - code - kiosk/print/print_test.go
- [[TestPDFGenerator()]] - code - kiosk/print/print_test.go
- [[Time_6]] - code - kiosk/print/pdf.go
- [[Warga_3]] - code - kiosk/print/pdf.go
- [[pdf.go]] - code - kiosk/print/pdf.go
- [[print_test.go]] - code - kiosk/print/print_test.go
- [[uuidString()]] - code - kiosk/print/pdf.go

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/Surat_Request_Handlers
SORT file.name ASC
```

## Connections to other communities
- 2 edges to [[_COMMUNITY_Kiosk API Tests]]

## Top bridge nodes
- [[NewPDFGenerator()]] - degree 5, connects to 1 community