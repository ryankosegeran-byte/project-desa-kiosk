# Graph Report - .  (2026-06-17)

## Corpus Check
- 280 files · ~211,234 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 858 nodes · 1334 edges · 89 communities (80 shown, 9 thin omitted)
- Extraction: 93% EXTRACTED · 7% INFERRED · 0% AMBIGUOUS · INFERRED: 92 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_API Helpers & Utilities|API Helpers & Utilities]]
- [[_COMMUNITY_Kiosk UI Components|Kiosk UI Components]]
- [[_COMMUNITY_Kiosk API Tests|Kiosk API Tests]]
- [[_COMMUNITY_Sync Handlers|Sync Handlers]]
- [[_COMMUNITY_JWT Authentication|JWT Authentication]]
- [[_COMMUNITY_Server Router & Middleware|Server Router & Middleware]]
- [[_COMMUNITY_Kiosk UI Dependencies|Kiosk UI Dependencies]]
- [[_COMMUNITY_Surat Request Handlers|Surat Request Handlers]]
- [[_COMMUNITY_Desa Repository|Desa Repository]]
- [[_COMMUNITY_Warga Repository|Warga Repository]]
- [[_COMMUNITY_Dashboard Dependencies|Dashboard Dependencies]]
- [[_COMMUNITY_Dashboard TSConfig|Dashboard TSConfig]]
- [[_COMMUNITY_Node TSConfig|Node TSConfig]]
- [[_COMMUNITY_JenisSurat Repository|JenisSurat Repository]]
- [[_COMMUNITY_Surat Repository|Surat Repository]]
- [[_COMMUNITY_OCR Provider Interface|OCR Provider Interface]]
- [[_COMMUNITY_Internal Models|Internal Models]]
- [[_COMMUNITY_OCR Google Gemini|OCR: Google Gemini]]
- [[_COMMUNITY_OCR Groq|OCR: Groq]]
- [[_COMMUNITY_OCR Mistral AI|OCR: Mistral AI]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 46|Community 46]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Community 48|Community 48]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_Community 50|Community 50]]
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
- [[_COMMUNITY_Community 62|Community 62]]

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
- `AuthMiddleware()` --calls--> `HandlerFunc`  [INFERRED]
  server/middleware/auth.go → kiosk/rfid/listener.go
- `KioskKeyMiddleware()` --calls--> `HandlerFunc`  [INFERRED]
  server/middleware/auth.go → kiosk/rfid/listener.go
- `RoleMiddleware()` --calls--> `HandlerFunc`  [INFERRED]
  server/middleware/auth.go → kiosk/rfid/listener.go
- `LoggerMiddleware()` --calls--> `HandlerFunc`  [INFERRED]
  server/api/router.go → kiosk/rfid/listener.go
- `main()` --calls--> `Duration`  [INFERRED]
  server/cmd/server/main.go → internal/auth/jwt.go

## Import Cycles
- None detected.

## Communities (89 total, 9 thin omitted)

### Community 0 - "API Helpers & Utilities"
Cohesion: 0.07
Nodes (31): parseJSON(), sendError(), sendJSON(), Claims, GetClaims(), Server, Request, ResponseWriter (+23 more)

### Community 1 - "Kiosk UI Components"
Cohesion: 0.07
Nodes (34): AddressPickerModal(), AddressPickerModalProps, FullKeyboard(), FullKeyboardProps, NIKInput(), NIKInputProps, RFIDStatus(), SelectOrInputModal() (+26 more)

### Community 2 - "Kiosk API Tests"
Cohesion: 0.07
Nodes (32): setupTestServer(), TestHandleCreateAndListSurat(), TestHandleGetWargaByNIK(), TestHandleGetWargaByRFID(), TestHandleListJenisSurat(), TestHandleRFIDEventsAndMock(), TestHandleStatus(), NewConfigRepository() (+24 more)

### Community 3 - "Sync Handlers"
Cohesion: 0.09
Nodes (31): HasRole(), AuthMiddleware(), GetKiosk(), KioskKeyMiddleware(), RoleMiddleware(), contextKey, Server, Request (+23 more)

### Community 4 - "JWT Authentication"
Cohesion: 0.10
Nodes (16): Claims, NewJWTManager(), JWTManager, SeedServerData(), NewUserRepository(), UserRepository, Duration, RegisteredClaims (+8 more)

### Community 5 - "Server Router & Middleware"
Cohesion: 0.11
Nodes (27): CORS(), sendError(), sendJSON(), setupStaticFileServer(), HandlerFunc, LoggerMiddleware(), NewServer(), Server (+19 more)

### Community 6 - "Kiosk UI Dependencies"
Cohesion: 0.07
Nodes (29): dependencies, leaflet, lucide-react, react, react-dom, react-router-dom, @types/leaflet, devDependencies (+21 more)

### Community 7 - "Surat Request Handlers"
Cohesion: 0.15
Nodes (13): Server, Request, ResponseWriter, Context, Time, Warga, T, FormatIndonesianDate() (+5 more)

### Community 8 - "Desa Repository"
Cohesion: 0.20
Nodes (8): NewDesaRepository(), DesaRepository, Desa, Context, DB, Kiosk, Row, Rows

### Community 9 - "Warga Repository"
Cohesion: 0.22
Nodes (8): Context, DB, NewWargaRepository(), WargaRepository, Row, Rows, Time, Warga

### Community 10 - "Dashboard Dependencies"
Cohesion: 0.10
Nodes (19): dependencies, astro, @astrojs/node, @astrojs/react, lucide-react, react, react-dom, @types/react (+11 more)

### Community 11 - "Dashboard TSConfig"
Cohesion: 0.11
Nodes (18): compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, jsx, lib, module, moduleDetection, moduleResolution (+10 more)

### Community 12 - "Node TSConfig"
Cohesion: 0.11
Nodes (17): compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, lib, module, moduleDetection, moduleResolution, noEmit (+9 more)

### Community 13 - "JenisSurat Repository"
Cohesion: 0.26
Nodes (7): Context, DB, NewJenisSuratRepository(), JenisSuratRepository, JenisSurat, Row, Rows

### Community 14 - "Surat Repository"
Cohesion: 0.24
Nodes (7): Context, DB, NewSuratRepository(), SuratRepository, Row, Rows, Surat

### Community 15 - "OCR Provider Interface"
Cohesion: 0.24
Nodes (8): MockProvider, ProviderStatus, Service, NewService(), OCRProvider, Context, KTPData, RWMutex

### Community 16 - "Internal Models"
Cohesion: 0.17
Nodes (15): JenisSurat, RawMessage, SuratTemplate, Time, Warga, Kiosk, KioskStatus, OCRProvider (+7 more)

### Community 17 - "OCR: Google Gemini"
Cohesion: 0.18
Nodes (10): NewGeminiProvider(), geminiContent, geminiGenerationConfig, geminiInlineData, geminiPart, GeminiProvider, geminiRequestPayload, geminiResponsePayload (+2 more)

### Community 18 - "OCR: Groq"
Cohesion: 0.18
Nodes (10): NewGroqProvider(), groqContentItem, groqImageURL, groqMessage, GroqProvider, groqRequestPayload, groqResponseFormat, groqResponsePayload (+2 more)

### Community 19 - "OCR: Mistral AI"
Cohesion: 0.18
Nodes (10): NewMistralProvider(), mistralContentItem, mistralImageURL, mistralMessage, MistralProvider, mistralRequestPayload, mistralResponseFormat, mistralResponsePayload (+2 more)

### Community 20 - "Community 20"
Cohesion: 0.27
Nodes (7): Context, DB, NewSuratRepository(), SuratRepository, Row, Rows, Surat

### Community 21 - "Community 21"
Cohesion: 0.29
Nodes (7): Context, DB, NewWargaRepository(), WargaRepository, Row, Rows, Warga

### Community 22 - "Community 22"
Cohesion: 0.23
Nodes (10): PdfReader, get_field_info(), get_full_annotation_field_id(), make_field_dict(), write_field_info(), fill_pdf_fields(), validation_error_for_field_value(), fill_pdf_form() (+2 more)

### Community 23 - "Community 23"
Cohesion: 0.28
Nodes (7): NewTemplateRepository(), TemplateRepository, Context, DB, Row, Rows, SuratTemplate

### Community 24 - "Community 24"
Cohesion: 0.17
Nodes (7): ../../../components/ActivityLogList, ../components/DashboardOverview, ../../../components/DesaManager, ../components/LoginForm, ../../components/WargaDraftComplete, ../styles/global.css, ../../layouts/DashboardLayout.astro

### Community 25 - "Community 25"
Cohesion: 0.21
Nodes (4): Surat, WargaFormData, ../../components/SuratTable, request()

### Community 26 - "Community 26"
Cohesion: 0.30
Nodes (6): Context, DB, NewJenisSuratRepository(), JenisSuratRepository, JenisSurat, SuratTemplate

### Community 27 - "Community 27"
Cohesion: 0.29
Nodes (9): Client, Config, ConfigRepository, Context, JenisSuratRepository, NomorSuratRepository, WargaRepository, Puller (+1 more)

### Community 28 - "Community 28"
Cohesion: 0.24
Nodes (8): ActivityLog, ActivityLogList(), Desa, DashboardOverview(), Stats, Desa, DesaManager(), getUser()

### Community 29 - "Community 29"
Cohesion: 0.33
Nodes (5): NewSyncRepository(), SyncRepository, Context, DB, SyncQueueItem

### Community 30 - "Community 30"
Cohesion: 0.42
Nodes (7): Detector, Config, Context, Puller, Pusher, Engine, NewEngine()

### Community 31 - "Community 31"
Cohesion: 0.25
Nodes (8): FieldDef, RawMessage, Time, DesaJenisSurat, FieldDef, FieldsSchema, JenisSurat, SuratTemplate

### Community 32 - "Community 32"
Cohesion: 0.36
Nodes (8): Client, Config, ConfigRepository, Context, SuratRepository, Pusher, NewPusher(), SyncRepository

### Community 33 - "Community 33"
Cohesion: 0.25
Nodes (5): OCRStatus, PROVIDER_LABELS, ProviderInfo, TestResult, ../../components/OCRProviderConfig

### Community 34 - "Community 34"
Cohesion: 0.29
Nodes (5): Desa, JenisSurat, Template, TemplatesList(), ../../components/TemplatesList

### Community 35 - "Community 35"
Cohesion: 0.29
Nodes (6): compilerOptions, jsx, jsxImportSource, exclude, extends, include

### Community 36 - "Community 36"
Cohesion: 0.43
Nodes (4): isDuplicateColumnErr(), Open(), Context, DB

### Community 37 - "Community 37"
Cohesion: 0.57
Nodes (6): setupTestDB(), TestJenisSuratRepository(), TestSuratRepository(), TestWargaRepository(), DB, T

### Community 38 - "Community 38"
Cohesion: 0.60
Nodes (3): Server, Request, ResponseWriter

### Community 39 - "Community 39"
Cohesion: 0.33
Nodes (4): Desa, JenisSurat, JenisSuratManager(), ../../../components/JenisSuratManager

### Community 40 - "Community 40"
Cohesion: 0.33
Nodes (4): Desa, Kiosk, KioskStatus(), ../../components/KioskStatus

### Community 41 - "Community 41"
Cohesion: 0.33
Nodes (3): Desa, User, ../../../components/UserManager

### Community 42 - "Community 42"
Cohesion: 0.33
Nodes (4): Desa, KTPData, WargaRegister(), ../../components/WargaRegister

### Community 43 - "Community 43"
Cohesion: 0.47
Nodes (6): Time, CreateUserRequest, LoginRequest, LoginResponse, User, UserActivityLog

### Community 44 - "Community 44"
Cohesion: 0.60
Nodes (3): Server, Request, ResponseWriter

### Community 46 - "Community 46"
Cohesion: 0.80
Nodes (5): getEnvBool(), Config, getEnv(), getEnvInt(), Load()

### Community 47 - "Community 47"
Cohesion: 0.70
Nodes (5): RawMessage, Time, CreateSuratRequest, NomorSuratBatch, Surat

### Community 48 - "Community 48"
Cohesion: 0.50
Nodes (3): Server, Request, ResponseWriter

### Community 50 - "Community 50"
Cohesion: 0.83
Nodes (3): hitl-loop.template.sh script, capture(), step()

### Community 52 - "Community 52"
Cohesion: 1.00
Nodes (4): Config, getEnv(), getEnvInt(), Load()

### Community 53 - "Community 53"
Cohesion: 0.67
Nodes (3): extract_form_structure(), main(), Extract form structure from a non-fillable PDF.  This script analyzes the PDF

### Community 54 - "Community 54"
Cohesion: 0.67
Nodes (3): is_server_ready(), main(), Wait for server to be ready by polling the port.

### Community 56 - "Community 56"
Cohesion: 1.00
Nodes (3): Time, KTPData, Warga

## Knowledge Gaps
- **191 isolated node(s):** `deploy.sh script`, `RegisteredClaims`, `Time`, `RawMessage`, `FieldDef` (+186 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **9 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.