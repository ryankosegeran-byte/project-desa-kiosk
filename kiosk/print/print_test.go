package print

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/project-desa-kiosk/internal/models"
)

func TestFormatIndonesianDate(t *testing.T) {
	testTime := time.Date(2026, time.June, 6, 10, 0, 0, 0, time.UTC)
	expected := "6 Juni 2026"
	got := FormatIndonesianDate(testTime)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}

	testTime2 := time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC)
	expected2 := "31 Desember 2026"
	got2 := FormatIndonesianDate(testTime2)
	if got2 != expected2 {
		t.Errorf("expected %q, got %q", expected2, got2)
	}
}

func TestPDFGenerator(t *testing.T) {
	// PDF generator will try to run Chrome, which might not be installed or configured in all test runners.
	// But we can verify it parses templates and handles errors.
	dir := filepath.Join("..", "data", "test_printed")
	defer os.RemoveAll(dir)

	g := NewPDFGenerator(dir)
	if g.OutputDir != dir {
		t.Errorf("expected output dir %q, got %q", dir, g.OutputDir)
	}

	// Test template error
	warga := &models.Warga{Nama: "Budi"}
	_, err := g.GeneratePDF(context.Background(), "invalid {{ template", warga, nil, "6 Juni 2026", "123/SKU/2026", "ALFRIDA", "19690206")
	if err == nil {
		t.Error("expected error parsing invalid template, got nil")
	}
}
