package docx

import (
	"archive/zip"
	"bytes"
	"reflect"
	"testing"
)

// makeDocx builds a minimal in-memory .docx (zip) from the given parts.
// word/document.xml is written first so iteration order is deterministic.
func makeDocx(t *testing.T, parts map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	order := []string{"word/document.xml", "word/header1.xml", "word/footer1.xml"}
	written := map[string]bool{}
	write := func(name string) {
		content, ok := parts[name]
		if !ok {
			return
		}
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
		written[name] = true
	}
	for _, n := range order {
		write(n)
	}
	for n := range parts {
		if !written[n] {
			write(n)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// Covers the three things that broke the old approach: split runs, multiple
// parts (header), and duplicate tokens.
func TestDetectTokens_SplitRunsHeaderAndDedup(t *testing.T) {
	doc := `<?xml version="1.0"?><w:document><w:body>` +
		`<w:p><w:r><w:t>Nama: {{na</w:t></w:r><w:r><w:t>ma}}</w:t></w:r></w:p>` + // split run
		`<w:p><w:r><w:t xml:space="preserve">Usaha: {{jenis_usaha}}</w:t></w:r></w:p>` +
		`<w:p><w:r><w:t>Lagi: {{nama}}</w:t></w:r></w:p>` + // duplicate -> deduped
		`</w:body></w:document>`
	header := `<?xml version="1.0"?><w:hdr><w:p><w:r><w:t>{{kepala_desa}}</w:t></w:r></w:p></w:hdr>`

	b := makeDocx(t, map[string]string{
		"word/document.xml": doc,
		"word/header1.xml":  header,
	})

	got, err := DetectTokens(b)
	if err != nil {
		t.Fatalf("DetectTokens error: %v", err)
	}
	want := []string{"nama", "jenis_usaha", "kepala_desa"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tokens mismatch:\n got=%v\nwant=%v", got, want)
	}
}

func TestDetectTokens_NoTokens(t *testing.T) {
	doc := `<?xml version="1.0"?><w:document><w:body>` +
		`<w:p><w:r><w:t>Tidak ada placeholder di sini.</w:t></w:r></w:p>` +
		`</w:body></w:document>`
	b := makeDocx(t, map[string]string{"word/document.xml": doc})

	got, err := DetectTokens(b)
	if err != nil {
		t.Fatalf("DetectTokens error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no tokens, got %v", got)
	}
}
