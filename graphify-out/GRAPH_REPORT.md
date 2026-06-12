# Graph Report - D:\PROJECT\MYPROJECT\project-desa-kiosk  (2026-06-12)

## Corpus Check
- 310 files · ~214,040 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1048 nodes · 1465 edges · 146 communities (108 shown, 38 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 92 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Authentication & Authorization|Authentication & Authorization]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Authentication & Authorization|Authentication & Authorization]]
- [[_COMMUNITY_StatusBar & Online Hooks|StatusBar & Online Hooks]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Authentication & Authorization|Authentication & Authorization]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_StatusBar & Online Hooks|StatusBar & Online Hooks]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Warga Registry & RFID Data|Warga Registry & RFID Data]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Warga Registry & RFID Data|Warga Registry & RFID Data]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Local-Server Sync Engine|Local-Server Sync Engine]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_StatusBar & Online Hooks|StatusBar & Online Hooks]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Warga Registry & RFID Data|Warga Registry & RFID Data]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Warga Registry & RFID Data|Warga Registry & RFID Data]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Warga Registry & RFID Data|Warga Registry & RFID Data]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_StatusBar & Online Hooks|StatusBar & Online Hooks]]
- [[_COMMUNITY_Community 51|Community 51]]
- [[_COMMUNITY_Community 52|Community 52]]
- [[_COMMUNITY_Community 53|Community 53]]
- [[_COMMUNITY_Community 54|Community 54]]
- [[_COMMUNITY_Community 55|Community 55]]
- [[_COMMUNITY_Community 56|Community 56]]
- [[_COMMUNITY_Community 57|Community 57]]
- [[_COMMUNITY_Community 58|Community 58]]
- [[_COMMUNITY_Community 59|Community 59]]
- [[_COMMUNITY_Community 60|Community 60]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Warga Registry & RFID Data|Warga Registry & RFID Data]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Community 64|Community 64]]
- [[_COMMUNITY_Community 65|Community 65]]
- [[_COMMUNITY_Warga Registry & RFID Data|Warga Registry & RFID Data]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Community 72|Community 72]]
- [[_COMMUNITY_OCR Document Parsing|OCR Document Parsing]]
- [[_COMMUNITY_Community 74|Community 74]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Surat & Document Generation|Surat & Document Generation]]
- [[_COMMUNITY_Community 77|Community 77]]
- [[_COMMUNITY_Community 78|Community 78]]
- [[_COMMUNITY_Community 79|Community 79]]
- [[_COMMUNITY_Community 80|Community 80]]
- [[_COMMUNITY_Community 81|Community 81]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]
- [[_COMMUNITY_Community 84|Community 84]]
- [[_COMMUNITY_Community 85|Community 85]]
- [[_COMMUNITY_Community 86|Community 86]]
- [[_COMMUNITY_Community 87|Community 87]]
- [[_COMMUNITY_Kiosk Backend Services|Kiosk Backend Services]]

## God Nodes (most connected - your core abstractions)
1. `GetClaims()` - 22 edges
2. `parseJSON()` - 19 edges
3. `DesaRepository` - 17 edges
4. `compilerOptions` - 17 edges
5. `WargaRepository` - 16 edges
6. `compilerOptions` - 16 edges
7. `setupTestServer()` - 15 edges
8. `getUser()` - 15 edges
9. `UserRepository` - 14 edges
10. `request()` - 14 edges

## Surprising Connections (you probably didn't know these)
- `RoleMiddleware()` --calls--> `HasRole()`  [INFERRED]
  server/middleware/auth.go → internal/auth/jwt.go
- `main()` --calls--> `Duration`  [INFERRED]
  server/cmd/server/main.go → internal/auth/jwt.go
- `main()` --calls--> `NewJWTManager()`  [INFERRED]
  server/cmd/server/main.go → internal/auth/jwt.go
- `AuthMiddleware()` --calls--> `HandlerFunc`  [INFERRED]
  server/middleware/auth.go → kiosk/rfid/listener.go
- `KioskKeyMiddleware()` --calls--> `HandlerFunc`  [INFERRED]
  server/middleware/auth.go → kiosk/rfid/listener.go

## Import Cycles
- None detected.

## Communities (146 total, 38 thin omitted)

### Community 1 - "Kiosk Backend Services"
Cohesion: 0.06
Nodes (34): parseJSON(), sendError(), sendJSON(), Claims, GetClaims(), Server, Request, ResponseWriter (+26 more)

### Community 2 - "Kiosk Backend Services"
Cohesion: 0.06
Nodes (49): CORS(), sendError(), sendJSON(), setupStaticFileServer(), HandlerFunc, LoggerMiddleware(), NewServer(), Server (+41 more)

### Community 3 - "Kiosk Backend Services"
Cohesion: 0.07
Nodes (34): AddressPickerModal(), AddressPickerModalProps, FullKeyboard(), FullKeyboardProps, NIKInput(), NIKInputProps, RFIDStatus(), SelectOrInputModal() (+26 more)

### Community 4 - "Surat & Document Generation"
Cohesion: 0.07
Nodes (29): setupTestServer(), TestHandleCreateAndListSurat(), TestHandleGetWargaByNIK(), TestHandleGetWargaByRFID(), TestHandleListJenisSurat(), TestHandleRFIDEventsAndMock(), TestHandleStatus(), NewConfigRepository() (+21 more)

### Community 5 - "Authentication & Authorization"
Cohesion: 0.07
Nodes (25): Claims, HasRole(), NewJWTManager(), JWTManager, SeedServerData(), Detector, Duration, Config (+17 more)

### Community 6 - "StatusBar & Online Hooks"
Cohesion: 0.07
Nodes (29): dependencies, leaflet, lucide-react, react, react-dom, react-router-dom, @types/leaflet, devDependencies (+21 more)

### Community 7 - "Surat & Document Generation"
Cohesion: 0.15
Nodes (13): Server, Request, ResponseWriter, Context, Time, Warga, T, FormatIndonesianDate() (+5 more)

### Community 8 - "Kiosk Backend Services"
Cohesion: 0.20
Nodes (8): NewDesaRepository(), DesaRepository, Desa, Context, DB, Kiosk, Row, Rows

### Community 9 - "Authentication & Authorization"
Cohesion: 0.22
Nodes (8): Context, DB, NewWargaRepository(), WargaRepository, Row, Rows, Time, Warga

### Community 10 - "Community 10"
Cohesion: 0.10
Nodes (19): dependencies, astro, @astrojs/node, @astrojs/react, lucide-react, react, react-dom, @types/react (+11 more)

### Community 11 - "Community 11"
Cohesion: 0.21
Nodes (8): NewUserRepository(), UserRepository, Context, DB, Row, Rows, User, UserActivityLog

### Community 12 - "Community 12"
Cohesion: 0.11
Nodes (18): compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, jsx, lib, module, moduleDetection, moduleResolution (+10 more)

### Community 13 - "Community 13"
Cohesion: 0.11
Nodes (17): compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, lib, module, moduleDetection, moduleResolution, noEmit (+9 more)

### Community 14 - "Surat & Document Generation"
Cohesion: 0.26
Nodes (7): Context, DB, NewJenisSuratRepository(), JenisSuratRepository, JenisSurat, Row, Rows

### Community 15 - "Surat & Document Generation"
Cohesion: 0.24
Nodes (7): Context, DB, NewSuratRepository(), SuratRepository, Row, Rows, Surat

### Community 16 - "StatusBar & Online Hooks"
Cohesion: 0.24
Nodes (7): MockProvider, ProviderStatus, NewService(), OCRProvider, Context, KTPData, RWMutex

### Community 17 - "Kiosk Backend Services"
Cohesion: 0.17
Nodes (15): JenisSurat, RawMessage, SuratTemplate, Time, Warga, Kiosk, KioskStatus, OCRProvider (+7 more)

### Community 18 - "Community 18"
Cohesion: 0.18
Nodes (10): NewGroqProvider(), groqContentItem, groqImageURL, groqMessage, GroqProvider, groqRequestPayload, groqResponseFormat, groqResponsePayload (+2 more)

### Community 19 - "Community 19"
Cohesion: 0.18
Nodes (10): NewMistralProvider(), mistralContentItem, mistralImageURL, mistralMessage, MistralProvider, mistralRequestPayload, mistralResponseFormat, mistralResponsePayload (+2 more)

### Community 20 - "Kiosk Backend Services"
Cohesion: 0.25
Nodes (8): GetKiosk(), Server, Request, ResponseWriter, Server, Request, ResponseWriter, Kiosk

### Community 21 - "Surat & Document Generation"
Cohesion: 0.27
Nodes (7): Context, DB, NewSuratRepository(), SuratRepository, Row, Rows, Surat

### Community 22 - "Warga Registry & RFID Data"
Cohesion: 0.29
Nodes (7): Context, DB, NewWargaRepository(), WargaRepository, Row, Rows, Warga

### Community 23 - "Community 23"
Cohesion: 0.23
Nodes (10): PdfReader, get_field_info(), get_full_annotation_field_id(), make_field_dict(), write_field_info(), fill_pdf_fields(), validation_error_for_field_value(), fill_pdf_form() (+2 more)

### Community 24 - "Surat & Document Generation"
Cohesion: 0.28
Nodes (7): NewTemplateRepository(), TemplateRepository, Context, DB, Row, Rows, SuratTemplate

### Community 25 - "Warga Registry & RFID Data"
Cohesion: 0.17
Nodes (6): ../../../components/ActivityLogList, ../components/DashboardOverview, ../../../components/DesaManager, ../components/LoginForm, ../../components/WargaDraftComplete, ../styles/global.css

### Community 26 - "Surat & Document Generation"
Cohesion: 0.21
Nodes (4): Surat, WargaFormData, ../../components/SuratTable, request()

### Community 27 - "Surat & Document Generation"
Cohesion: 0.30
Nodes (6): Context, DB, NewJenisSuratRepository(), JenisSuratRepository, JenisSurat, SuratTemplate

### Community 28 - "Surat & Document Generation"
Cohesion: 0.29
Nodes (8): Client, Config, ConfigRepository, Context, JenisSuratRepository, NomorSuratRepository, WargaRepository, NewPuller()

### Community 29 - "Community 29"
Cohesion: 0.24
Nodes (8): ActivityLog, ActivityLogList(), Desa, DashboardOverview(), Stats, Desa, DesaManager(), getUser()

### Community 30 - "Local-Server Sync Engine"
Cohesion: 0.33
Nodes (5): NewSyncRepository(), SyncRepository, Context, DB, SyncQueueItem

### Community 31 - "Surat & Document Generation"
Cohesion: 0.25
Nodes (8): FieldDef, RawMessage, Time, DesaJenisSurat, FieldDef, FieldsSchema, JenisSurat, SuratTemplate

### Community 32 - "Surat & Document Generation"
Cohesion: 0.36
Nodes (7): Client, Config, ConfigRepository, Context, SuratRepository, NewPusher(), SyncRepository

### Community 33 - "StatusBar & Online Hooks"
Cohesion: 0.25
Nodes (5): OCRStatus, PROVIDER_LABELS, ProviderInfo, TestResult, ../../components/OCRProviderConfig

### Community 34 - "Surat & Document Generation"
Cohesion: 0.29
Nodes (5): Desa, JenisSurat, Template, TemplatesList(), ../../components/TemplatesList

### Community 35 - "Community 35"
Cohesion: 0.29
Nodes (6): compilerOptions, jsx, jsxImportSource, exclude, extends, include

### Community 36 - "Community 36"
Cohesion: 0.43
Nodes (4): isDuplicateColumnErr(), Open(), Context, DB

### Community 37 - "Surat & Document Generation"
Cohesion: 0.57
Nodes (6): setupTestDB(), TestJenisSuratRepository(), TestSuratRepository(), TestWargaRepository(), DB, T

### Community 39 - "Warga Registry & RFID Data"
Cohesion: 0.60
Nodes (3): Server, Request, ResponseWriter

### Community 40 - "Surat & Document Generation"
Cohesion: 0.33
Nodes (4): Desa, JenisSurat, JenisSuratManager(), ../../../components/JenisSuratManager

### Community 41 - "Kiosk Backend Services"
Cohesion: 0.33
Nodes (4): Desa, Kiosk, KioskStatus(), ../../components/KioskStatus

### Community 42 - "Community 42"
Cohesion: 0.33
Nodes (3): Desa, User, ../../../components/UserManager

### Community 43 - "Warga Registry & RFID Data"
Cohesion: 0.33
Nodes (4): Desa, KTPData, WargaRegister(), ../../components/WargaRegister

### Community 45 - "Community 45"
Cohesion: 0.47
Nodes (5): Time, CreateUserRequest, LoginRequest, LoginResponse, UserActivityLog

### Community 47 - "Community 47"
Cohesion: 0.80
Nodes (4): getEnvBool(), getEnv(), getEnvInt(), Load()

### Community 48 - "Surat & Document Generation"
Cohesion: 0.70
Nodes (4): RawMessage, Time, CreateSuratRequest, NomorSuratBatch

### Community 50 - "StatusBar & Online Hooks"
Cohesion: 0.50
Nodes (3): Server, Request, ResponseWriter

### Community 53 - "Community 53"
Cohesion: 0.83
Nodes (3): hitl-loop.template.sh script, capture(), step()

### Community 55 - "Community 55"
Cohesion: 1.00
Nodes (3): getEnv(), getEnvInt(), Load()

### Community 58 - "Community 58"
Cohesion: 0.67
Nodes (3): extract_form_structure(), main(), Extract form structure from a non-fillable PDF.  This script analyzes the PDF

### Community 59 - "Community 59"
Cohesion: 0.67
Nodes (3): is_server_ready(), main(), Wait for server to be ready by polling the port.

## Knowledge Gaps
- **191 isolated node(s):** `deploy.sh script`, `RegisteredClaims`, `Time`, `RawMessage`, `FieldDef` (+186 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **38 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.