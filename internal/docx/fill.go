package docx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	reParaBlock = regexp.MustCompile(`(?s)<w:p[ >][\s\S]*?</w:p>`)
	reRunText   = regexp.MustCompile(`(?s)(<w:t(?:\s[^>]*)?>)([\s\S]*?)(</w:t>)`)
)

// FillDocx fills {{token}} placeholders in a DOCX and returns new DOCX bytes.
// Works at paragraph level: text across all runs is concatenated before matching,
// so split-run tokens (Word breaks {{ and }} into separate XML runs) are handled.
// Pure Go — no Word or LibreOffice required.
func FillDocx(docxBytes []byte, values map[string]string) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		return nil, fmt.Errorf("buka docx zip: %w", err)
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("buka entry %s: %w", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("baca entry %s: %w", f.Name, err)
		}

		if isTextPart(f.Name) {
			data = applyTokens(data, values)
		}

		fw, err := w.Create(f.Name)
		if err != nil {
			return nil, err
		}
		if _, err = fw.Write(data); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func isTextPart(name string) bool {
	base := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		base = name[i+1:]
	}
	return base == "document.xml" ||
		strings.HasPrefix(base, "header") ||
		strings.HasPrefix(base, "footer")
}

func applyTokens(data []byte, values map[string]string) []byte {
	return reParaBlock.ReplaceAllFunc(data, func(para []byte) []byte {
		return fillParagraph(para, values)
	})
}

// fillParagraph concatenates text from all <w:t> runs in a paragraph,
// replaces tokens, then writes the full result into the first run and
// clears the rest. Paragraphs without tokens are returned unchanged.
func fillParagraph(para []byte, values map[string]string) []byte {
	var flat strings.Builder
	reRunText.ReplaceAllFunc(para, func(m []byte) []byte {
		parts := reRunText.FindSubmatch(m)
		if parts != nil {
			flat.Write(parts[2])
		}
		return m
	})
	combined := flat.String()

	if !strings.Contains(combined, "{{") {
		return para
	}

	filled := combined
	for k, v := range values {
		filled = strings.ReplaceAll(filled, "{{"+k+"}}", v)
	}
	if filled == combined {
		return para
	}

	n := 0
	return reRunText.ReplaceAllFunc(para, func(m []byte) []byte {
		parts := reRunText.FindSubmatch(m)
		if parts == nil {
			return m
		}
		n++
		open, close_ := string(parts[1]), string(parts[3])
		if n == 1 {
			if !strings.Contains(open, "xml:space") {
				open = strings.TrimSuffix(open, ">") + ` xml:space="preserve">`
			}
			return []byte(open + xmlEsc(filled) + close_)
		}
		return []byte(open + close_)
	})
}

func xmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
