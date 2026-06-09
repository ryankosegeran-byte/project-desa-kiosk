package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/project-desa-kiosk/internal/models"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	// Open in-memory SQLite database
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("gagal membuka database test: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestWargaRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewWargaRepository(db)
	ctx := context.Background()
	desaID := "desa-test-uuid"

	w := &models.Warga{
		ID:              "warga-1",
		NIK:             "1234567890123456",
		RFIDUID:         "RFID-ABCD-1234",
		Nama:            "Ahmad Fauzi",
		TempatLahir:     "Jakarta",
		TanggalLahir:    "1990-01-01",
		JenisKelamin:    "L",
		Alamat:          "Jl. Sudirman No. 1",
		RT:              "01",
		RW:              "01",
		Kelurahan:       "Senayan",
		Kecamatan:       "Kebayoran Baru",
		Kabupaten:       "Jakarta Selatan",
		Provinsi:        "DKI Jakarta",
		Agama:           "Islam",
		StatusKawin:     "Belum Kawin",
		Pekerjaan:       "PNS",
		Kewarganegaraan: "WNI",
		DesaID:          desaID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 1. Test Upsert (Insert)
	err := repo.Upsert(ctx, w)
	if err != nil {
		t.Fatalf("gagal upsert warga: %v", err)
	}

	// 2. Test FindByRFID
	found, err := repo.FindByRFID(ctx, "RFID-ABCD-1234")
	if err != nil {
		t.Fatalf("gagal mencari warga by RFID: %v", err)
	}
	if found.Nama != w.Nama {
		t.Errorf("expected nama %s, got %s", w.Nama, found.Nama)
	}

	// Test Case-insensitive RFID lookup
	foundLower, err := repo.FindByRFID(ctx, "rfid-abcd-1234")
	if err != nil {
		t.Fatalf("gagal mencari warga by lowercase RFID: %v", err)
	}
	if foundLower.ID != w.ID {
		t.Errorf("expected ID %s, got %s", w.ID, foundLower.ID)
	}

	// 3. Test FindByNIK
	foundByNIK, err := repo.FindByNIK(ctx, "1234567890123456")
	if err != nil {
		t.Fatalf("gagal mencari warga by NIK: %v", err)
	}
	if foundByNIK.Nama != w.Nama {
		t.Errorf("expected nama %s, got %s", w.Nama, foundByNIK.Nama)
	}

	// 4. Test Search
	results, err := repo.Search(ctx, "Ahmad")
	if err != nil {
		t.Fatalf("gagal search warga: %v", err)
	}
	if len(results) != 1 || results[0].Nama != w.Nama {
		t.Errorf("search result tidak sesuai")
	}

	// 5. Test Upsert (Update)
	w.Nama = "Ahmad Fauzi Updated"
	err = repo.Upsert(ctx, w)
	if err != nil {
		t.Fatalf("gagal update warga via upsert: %v", err)
	}

	foundUpdated, _ := repo.FindByNIK(ctx, "1234567890123456")
	if foundUpdated.Nama != "Ahmad Fauzi Updated" {
		t.Errorf("expected updated nama %s, got %s", "Ahmad Fauzi Updated", foundUpdated.Nama)
	}

	// 6. Test Find not found
	_, err = repo.FindByRFID(ctx, "RFID-NON-EXISTENT")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestJenisSuratRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewJenisSuratRepository(db)
	ctx := context.Background()

	js := &models.JenisSurat{
		ID:           "js-1",
		Kode:         "SK_DOMISILI",
		Nama:         "Surat Keterangan Domisili",
		Deskripsi:    "Keterangan domisili",
		FieldsSchema: json.RawMessage(`{"fields":[]}`),
		Aktif:        true,
		Urutan:       1,
		UpdatedAt:    time.Now(),
	}

	// 1. Test Upsert
	err := repo.Upsert(ctx, js)
	if err != nil {
		t.Fatalf("gagal upsert jenis_surat: %v", err)
	}

	// 2. Test FindByID
	found, err := repo.FindByID(ctx, "js-1")
	if err != nil {
		t.Fatalf("gagal mencari jenis_surat: %v", err)
	}
	if found.Kode != js.Kode {
		t.Errorf("expected kode %s, got %s", js.Kode, found.Kode)
	}

	// 3. Test ListAktif
	list, err := repo.ListAktif(ctx)
	if err != nil {
		t.Fatalf("gagal list jenis_surat aktif: %v", err)
	}
	if len(list) != 1 || list[0].Kode != js.Kode {
		t.Errorf("expected active list size 1 with kode %s, got %v", js.Kode, list)
	}

	// 4. Test Template Upsert & Get
	tpl := &models.SuratTemplate{
		ID:           "tpl-1",
		JenisSuratID: "js-1",
		DesaID:       "desa-1",
		TemplateHTML: "<html><body>{{.Warga.Nama}}</body></html>",
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err = repo.UpsertTemplate(ctx, tpl)
	if err != nil {
		t.Fatalf("gagal upsert template: %v", err)
	}

	foundTpl, err := repo.GetTemplate(ctx, "js-1", "desa-1")
	if err != nil {
		t.Fatalf("gagal get template: %v", err)
	}
	if foundTpl.TemplateHTML != tpl.TemplateHTML {
		t.Errorf("expected template html %s, got %s", tpl.TemplateHTML, foundTpl.TemplateHTML)
	}
}

func TestSuratRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSuratRepository(db)
	ctx := context.Background()
	desaID := "desa-1"

	// Prerequisite data to satisfy foreign key constraints
	jsRepo := NewJenisSuratRepository(db)
	wargaRepo := NewWargaRepository(db)

	js := &models.JenisSurat{
		ID:           "js-1",
		Kode:         "SK_DOMISILI",
		Nama:         "Surat Keterangan Domisili",
		FieldsSchema: json.RawMessage(`{"fields":[]}`),
		Aktif:        true,
		UpdatedAt:    time.Now(),
	}
	if err := jsRepo.Upsert(ctx, js); err != nil {
		t.Fatalf("gagal upsert prerequisite jenis_surat: %v", err)
	}

	w := &models.Warga{
		ID:              "warga-1",
		NIK:             "1234567890123456",
		Nama:            "Ahmad",
		JenisKelamin:    "L",
		DesaID:          desaID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := wargaRepo.Upsert(ctx, w); err != nil {
		t.Fatalf("gagal upsert prerequisite warga: %v", err)
	}

	s := &models.Surat{
		ID:             "surat-1",
		JenisSuratID:   "js-1",
		JenisSuratKode: "SK_DOMISILI",
		JenisSuratNama: "Surat Keterangan Domisili",
		WargaID:        "warga-1",
		NIKPemohon:     "1234567890123456",
		NamaPemohon:    "Ahmad",
		DataSurat:      json.RawMessage(`{"tujuan":"test"}`),
		Status:         models.SuratStatusDraft,
		DesaID:         desaID,
		CreatedAt:      time.Now(),
		Synced:         false,
	}

	// 1. Test Create
	err := repo.Create(ctx, s)
	if err != nil {
		t.Fatalf("gagal membuat surat: %v", err)
	}

	// 2. Test FindByID
	found, err := repo.FindByID(ctx, "surat-1")
	if err != nil {
		t.Fatalf("gagal mencari surat: %v", err)
	}
	if found.NamaPemohon != s.NamaPemohon {
		t.Errorf("expected nama pemohon %s, got %s", s.NamaPemohon, found.NamaPemohon)
	}
	if found.Status != models.SuratStatusDraft {
		t.Errorf("expected status DRAFT, got %s", found.Status)
	}

	// 3. Test MarkPrinted (updates status and queues in sync_queue)
	err = repo.MarkPrinted(ctx, "surat-1", "printed/surat-1.pdf")
	if err != nil {
		t.Fatalf("gagal mark printed: %v", err)
	}

	foundPrinted, err := repo.FindByID(ctx, "surat-1")
	if err != nil {
		t.Fatalf("gagal mencari surat printed: %v", err)
	}
	if foundPrinted.Status != models.SuratStatusPrinted {
		t.Errorf("expected status PRINTED, got %s", foundPrinted.Status)
	}
	if foundPrinted.PrintedAt == nil {
		t.Errorf("expected printed_at to be set")
	}

	// Verify sync queue record was created
	var queueCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sync_queue WHERE entity_id = ?", "surat-1").Scan(&queueCount)
	if err != nil {
		t.Fatalf("gagal query sync_queue: %v", err)
	}
	if queueCount != 1 {
		t.Errorf("expected 1 record in sync_queue, got %d", queueCount)
	}

	// 4. Test ListUnsynced
	unsynced, err := repo.ListUnsynced(ctx, desaID)
	if err != nil {
		t.Fatalf("gagal list unsynced: %v", err)
	}
	if len(unsynced) != 1 || unsynced[0].ID != "surat-1" {
		t.Errorf("expected 1 unsynced record, got %v", unsynced)
	}

	// 5. Test MarkSynced
	err = repo.MarkSynced(ctx, "surat-1")
	if err != nil {
		t.Fatalf("gagal mark synced: %v", err)
	}

	foundSynced, _ := repo.FindByID(ctx, "surat-1")
	if foundSynced.Status != models.SuratStatusSynced || !foundSynced.Synced || foundSynced.SyncedAt == nil {
		t.Errorf("status sync tidak ter-update dengan benar: %+v", foundSynced)
	}

	// 6. Test ListToday
	todayList, err := repo.ListToday(ctx, desaID)
	if err != nil {
		t.Fatalf("gagal list today: %v", err)
	}
	if len(todayList) != 1 || todayList[0].ID != "surat-1" {
		t.Errorf("expected 1 today record, got %v", todayList)
	}
}
