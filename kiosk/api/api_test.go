package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/kiosk/config"
	"github.com/project-desa-kiosk/kiosk/db"
	"github.com/project-desa-kiosk/kiosk/print"
	"github.com/project-desa-kiosk/kiosk/rfid"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	// 1. Open in-memory SQLite DB
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("gagal membuka database test: %v", err)
	}

	// 2. Repositories
	wargaRepo := db.NewWargaRepository(database)
	suratRepo := db.NewSuratRepository(database)
	jenisSuratRepo := db.NewJenisSuratRepository(database)
	configRepo := db.NewConfigRepository(database)
	nomorSuratRepo := db.NewNomorSuratRepository(database)

	desaID := "desa-test-uuid"
	// 3. Seed data
	if err := db.SeedLocalData(database, desaID); err != nil {
		database.Close()
		t.Fatalf("gagal seed database: %v", err)
	}

	// 4. Config
	cfg := &config.Config{
		DesaID:    desaID,
		KioskName: "Kiosk Test",
	}

	rfidBroker := rfid.NewBroker()
	pdfGen := print.NewPDFGenerator("data/printed_test")
	printer := print.NewPrinter("")
	docxRenderer := print.NewDocxRenderer("data/printed_test")

	server := NewServer(cfg, wargaRepo, suratRepo, jenisSuratRepo, configRepo, nomorSuratRepo, rfidBroker, rfid.NewSessionWatcher("", ""), pdfGen, printer, docxRenderer)

	cleanup := func() {
		database.Close()
	}

	return server, cleanup
}

func TestHandleStatus(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req, err := http.NewRequest("GET", "/api/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status 200, got %v", status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response["desa_id"] != "desa-test-uuid" {
		t.Errorf("expected desa_id 'desa-test-uuid', got %v", response["desa_id"])
	}
	if response["kiosk_name"] != "Kiosk Test" {
		t.Errorf("expected kiosk_name 'Kiosk Test', got %v", response["kiosk_name"])
	}
}

func TestHandleGetWargaByRFID(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Test existing RFID UID (seeded as "1234567890")
	req, _ := http.NewRequest("GET", "/api/warga/rfid/1234567890", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %v", rr.Code)
	}

	var w models.Warga
	if err := json.Unmarshal(rr.Body.Bytes(), &w); err != nil {
		t.Fatal(err)
	}
	if w.Nama != "Budi Santoso" {
		t.Errorf("expected warga 'Budi Santoso', got '%s'", w.Nama)
	}

	// 2. Test nonexistent RFID UID
	req, _ = http.NewRequest("GET", "/api/warga/rfid/nonexistent", nil)
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %v", rr.Code)
	}
}

func TestHandleGetWargaByNIK(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Test existing NIK (seeded as "3201234567890001")
	req, _ := http.NewRequest("GET", "/api/warga/nik/3201234567890001", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %v", rr.Code)
	}

	var w models.Warga
	if err := json.Unmarshal(rr.Body.Bytes(), &w); err != nil {
		t.Fatal(err)
	}
	if w.Nama != "Budi Santoso" {
		t.Errorf("expected warga 'Budi Santoso', got '%s'", w.Nama)
	}

	// 2. Test invalid NIK length
	req, _ = http.NewRequest("GET", "/api/warga/nik/12345", nil)
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %v", rr.Code)
	}
}

func TestHandleListJenisSurat(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", "/api/jenis-surat", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %v", rr.Code)
	}

	var list []models.JenisSurat
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}

	// Seeded data has 12 items
	if len(list) != 12 {
		t.Errorf("expected 12 jenis surat, got %d", len(list))
	}
}

func TestHandleCreateAndListSurat(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. List active jenis_surat to get an ID
	reqJS, _ := http.NewRequest("GET", "/api/jenis-surat", nil)
	rrJS := httptest.NewRecorder()
	server.Handler().ServeHTTP(rrJS, reqJS)

	var list []models.JenisSurat
	json.Unmarshal(rrJS.Body.Bytes(), &list)
	if len(list) == 0 {
		t.Fatal("no jenis surat seeded")
	}
	targetJS := list[0] // e.g. SK_DOMISILI

	// 1.5. Lookup Budi Santoso to get his real UUID
	reqWarga, _ := http.NewRequest("GET", "/api/warga/nik/3201234567890001", nil)
	rrWarga := httptest.NewRecorder()
	server.Handler().ServeHTTP(rrWarga, reqWarga)
	if rrWarga.Code != http.StatusOK {
		t.Fatalf("gagal mencari warga test untuk create surat: %v", rrWarga.Body.String())
	}
	var testWarga models.Warga
	if err := json.Unmarshal(rrWarga.Body.Bytes(), &testWarga); err != nil {
		t.Fatal(err)
	}

	// 2. Submit a create request
	createReq := models.CreateSuratRequest{
		JenisSuratID: targetJS.ID,
		WargaID:      testWarga.ID,
		NIKPemohon:   "3201234567890001",
		NamaPemohon:  "Budi Santoso",
		DataSurat:    json.RawMessage(`{"tujuan":"Membuka Rekening Bank"}`),
	}

	bodyBytes, _ := json.Marshal(createReq)
	reqCreate, _ := http.NewRequest("POST", "/api/surat", bytes.NewBuffer(bodyBytes))
	reqCreate.Header.Set("Content-Type", "application/json")

	rrCreate := httptest.NewRecorder()
	server.Handler().ServeHTTP(rrCreate, reqCreate)

	if rrCreate.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %v, body: %s", rrCreate.Code, rrCreate.Body.String())
	}

	var createdSurat models.Surat
	if err := json.Unmarshal(rrCreate.Body.Bytes(), &createdSurat); err != nil {
		t.Fatal(err)
	}
	if createdSurat.ID == "" {
		t.Error("expected generated ID for created surat")
	}
	if createdSurat.JenisSuratKode != targetJS.Kode {
		t.Errorf("expected denormalized kode %s, got %s", targetJS.Kode, createdSurat.JenisSuratKode)
	}

	// 3. Test list today's surat
	reqList, _ := http.NewRequest("GET", "/api/surat", nil)
	rrList := httptest.NewRecorder()
	server.Handler().ServeHTTP(rrList, reqList)

	if rrList.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %v", rrList.Code)
	}

	var todayList []models.Surat
	if err := json.Unmarshal(rrList.Body.Bytes(), &todayList); err != nil {
		t.Fatal(err)
	}
	if len(todayList) != 1 || todayList[0].ID != createdSurat.ID {
		t.Errorf("expected 1 today surat record, got %v", todayList)
	}
}

func TestHandleRFIDEventsAndMock(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Start broker background process
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server.rfidBroker.Start(ctx)

	// Since SSE stream is long-running and blocking, we can test the mock scan publishing directly.
	// We'll subscribe to the broker manually, trigger the mock scan POST endpoint, and verify we receive the event.
	events := server.rfidBroker.Subscribe()
	defer server.rfidBroker.Unsubscribe(events)

	// 1. Post to mock scan endpoint
	mockPayload := map[string]string{"uid": "RFID-TEST-999"}
	bodyBytes, _ := json.Marshal(mockPayload)
	reqMock, _ := http.NewRequest("POST", "/api/rfid/mock", bytes.NewBuffer(bodyBytes))
	reqMock.Header.Set("Content-Type", "application/json")

	rrMock := httptest.NewRecorder()
	server.Handler().ServeHTTP(rrMock, reqMock)

	if rrMock.Code != http.StatusOK {
		t.Errorf("expected 200 OK for mock trigger, got %v", rrMock.Code)
	}

	// 2. Read from our manual subscription channel to verify event was broadcast
	select {
	case uid := <-events:
		if uid != "RFID-TEST-999" {
			t.Errorf("expected received UID 'RFID-TEST-999', got '%s'", uid)
		}
	case <-time.After(time.Second * 2):
		t.Error("timed out waiting for RFID broadcast event")
	}
}
