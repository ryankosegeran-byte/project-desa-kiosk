//go:build windows

package print

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Renders the real Kalawat SKU .docx through the Word engine end-to-end and
// checks a sizable PDF is produced (i.e. the kop/logo survived). Skips when the
// sample file or Microsoft Word is unavailable.
func TestDocxRenderer_RealFile(t *testing.T) {
	src := filepath.Join("..", "..", "template_surat", "kalawat", "surat_keterangan_usaha", "sku.docx")
	docxBytes, err := os.ReadFile(src)
	if err != nil {
		t.Skipf("file contoh tidak tersedia: %v", err)
	}

	r := NewDocxRenderer(t.TempDir())
	defer r.Close()

	// First render cold-starts Word.
	p1, err := r.Render(docxBytes, map[string]string{})
	if err != nil {
		if strings.Contains(err.Error(), "Microsoft Word") {
			t.Skipf("Word tidak tersedia: %v", err)
		}
		t.Fatalf("render pertama gagal: %v", err)
	}
	// Second render must reuse the warm Word instance.
	p2, err := r.Render(docxBytes, map[string]string{})
	if err != nil {
		t.Fatalf("render kedua (warm) gagal: %v", err)
	}
	if p1 == p2 {
		t.Fatalf("kedua render menghasilkan path sama: %s", p1)
	}

	for _, p := range []string{p1, p2} {
		fi, err := os.Stat(p)
		if err != nil {
			t.Fatalf("PDF tidak ditemukan: %v", err)
		}
		if fi.Size() < 50_000 {
			t.Fatalf("PDF terlalu kecil (%d bytes) — kemungkinan kop/logo hilang", fi.Size())
		}
	}
	t.Logf("OK: 2 render warm-instance berhasil")
}
