// Package docx provides lightweight helpers for working with .docx (Office Open
// XML) files without a full Office engine. It is used at authoring/upload time
// to discover the {{token}} placeholders an admin has marked inside a template.
//
// Rendering a filled letter (DOCX -> PDF) is NOT done here; that requires a real
// Office engine (Word/LibreOffice) and lives in the kiosk print package.
package docx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var (
	// reText captures the text inside each <w:t> ... </w:t> run.
	reText = regexp.MustCompile(`(?s)<w:t[^>]*>(.*?)</w:t>`)
	// reToken matches a {{snake_case}} or {{with-dashes}} placeholder and captures its key.
	reToken = regexp.MustCompile(`\{\{([a-z0-9_-]+)\}\}`)
)

// DetectTokens scans a DOCX and returns the distinct {{snake_case}} placeholder
// keys (without braces) found in the document body, headers and footers, in
// first-seen order.
//
// Word often splits a typed token like "{{nama}}" across several runs in the
// XML (e.g. "{{na" + "ma}}"). To handle that, the text of all <w:t> runs in a
// part is concatenated before matching, so split tokens are rejoined. No
// modification of the DOCX is performed.
func DetectTokens(docxBytes []byte) ([]string, error) {
	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		return nil, fmt.Errorf("gagal membaca docx (zip): %w", err)
	}

	seen := map[string]bool{}
	var tokens []string

	for _, f := range zr.File {
		if !isContentPart(f.Name) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("gagal membuka %s: %w", f.Name, err)
		}
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("gagal membaca %s: %w", f.Name, err)
		}

		// Concatenate all run text so tokens split across runs are rejoined.
		var sb strings.Builder
		for _, m := range reText.FindAllStringSubmatch(buf.String(), -1) {
			sb.WriteString(m[1])
		}

		for _, m := range reToken.FindAllStringSubmatch(sb.String(), -1) {
			key := m[1]
			if !seen[key] {
				seen[key] = true
				tokens = append(tokens, key)
			}
		}
	}

	return tokens, nil
}

// isContentPart reports whether an OOXML part can contain visible text that may
// hold placeholders: the main document body, headers, and footers.
func isContentPart(name string) bool {
	if name == "word/document.xml" {
		return true
	}
	if strings.HasPrefix(name, "word/header") && strings.HasSuffix(name, ".xml") {
		return true
	}
	if strings.HasPrefix(name, "word/footer") && strings.HasSuffix(name, ".xml") {
		return true
	}
	return false
}
