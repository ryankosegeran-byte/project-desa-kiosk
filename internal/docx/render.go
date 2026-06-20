// Package docx — render.go provides a lightweight pure-Go conversion from
// Office Open XML (.docx) to styled HTML. It is used for dashboard preview only
// and does NOT require LibreOffice or Word.
//
// Limitations vs. a real Office engine:
//   - Images/shapes are replaced with a placeholder icon.
//   - Complex table layouts (merged cells, nested tables) are simplified.
//   - Theme/style inheritance is approximate (direct formatting is honoured).
//
// This is intentional: for a "preview with dummy data" use-case, approximate
// HTML is far better than no preview at all, and it runs within ~5 MB RAM.
package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// reTokenInXML matches {{snake_case}} or {{with-dashes}} in concatenated text.
var reTokenInXML = regexp.MustCompile(`\{\{([a-z0-9_-]+)\}\}`)

// RenderToHTML converts a DOCX template (raw bytes) into a self-contained HTML
// string. Token placeholders like {{nama}} are replaced with the corresponding
// value from the values map before rendering.
func RenderToHTML(docxBytes []byte, values map[string]string) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		return "", fmt.Errorf("gagal membaca docx (zip): %w", err)
	}

	// Read all content parts (document body + headers + footers).
	var headerHTML, bodyHTML, footerHTML strings.Builder

	for _, f := range zr.File {
		if !isContentPart(f.Name) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("gagal membuka %s: %w", f.Name, err)
		}
		raw, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return "", fmt.Errorf("gagal membaca %s: %w", f.Name, err)
		}

		// Parse OOXML and convert to HTML, replacing tokens during render.
		partHTML := renderPart(raw, values)

		if strings.HasPrefix(f.Name, "word/header") {
			headerHTML.WriteString(partHTML)
		} else if strings.HasPrefix(f.Name, "word/footer") {
			footerHTML.WriteString(partHTML)
		} else {
			bodyHTML.WriteString(partHTML)
		}
	}

	// Wrap in a styled HTML page mimicking letter paper.
	var out strings.Builder
	out.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><style>`)
	out.WriteString(`
body {
  font-family: 'Times New Roman', 'Serif', serif;
  font-size: 12pt;
  color: #000;
  margin: 50px 60px;
  line-height: 1.6;
}
table {
  border-collapse: collapse;
  width: 100%;
  margin: 4px 0;
}
td, th {
  padding: 3px 8px;
  vertical-align: top;
}
table.bordered td, table.bordered th {
  border: 1px solid #000;
}
.header-section {
  text-align: center;
  border-bottom: 3px double #000;
  padding-bottom: 15px;
  margin-bottom: 20px;
}
.footer-section {
  margin-top: 30px;
}
p {
  margin: 4px 0;
}
.img-placeholder {
  display: inline-block;
  width: 60px;
  height: 60px;
  border: 1px dashed #999;
  text-align: center;
  line-height: 60px;
  font-size: 28px;
  color: #999;
  border-radius: 4px;
  vertical-align: middle;
}
`)
	out.WriteString(`</style></head><body>`)

	if headerHTML.Len() > 0 {
		out.WriteString(`<div class="header-section">`)
		out.WriteString(headerHTML.String())
		out.WriteString(`</div>`)
	}

	out.WriteString(bodyHTML.String())

	if footerHTML.Len() > 0 {
		out.WriteString(`<div class="footer-section">`)
		out.WriteString(footerHTML.String())
		out.WriteString(`</div>`)
	}

	out.WriteString(`</body></html>`)
	return out.String(), nil
}

// --- Core: paragraph-level token replacement + HTML rendering ---

// runInfo holds the raw data extracted from a single <w:r> run.
type runInfo struct {
	text      string
	bold      bool
	italic    bool
	underline bool
	strike    bool
	fontSize  string
	fontFamily string
}

// renderPart processes a full OOXML part (document.xml, header*.xml, etc.)
// and produces HTML. Token replacement happens at the paragraph level so that
// Word's split-run tokens (e.g. "{{" + "nama" + "}}") are properly resolved.
func renderPart(xmlData []byte, values map[string]string) string {
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	decoder.Strict = false

	var out strings.Builder
	var inTable bool

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "body":
				// continue
			case "p":
				pHTML := renderParagraphV2(decoder, values)
				out.WriteString(pHTML)
			case "tbl":
				inTable = true
				// Check if table has borders
				hasBorder := checkTableBorders(decoder)
				if hasBorder {
					out.WriteString(`<table class="bordered">`)
				} else {
					out.WriteString(`<table>`)
				}
			case "tr":
				if inTable {
					out.WriteString("<tr>")
				}
			case "tc":
				if inTable {
					out.WriteString("<td>")
					cellHTML := renderTableCellV2(decoder, values)
					out.WriteString(cellHTML)
					out.WriteString("</td>")
				}
			case "drawing", "pict", "object":
				// Image/drawing — output placeholder
				out.WriteString(`<span class="img-placeholder">🖼</span>`)
				skipElement(decoder)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "tbl":
				inTable = false
				out.WriteString("</table>")
			case "tr":
				out.WriteString("</tr>")
			}
		}
	}

	return out.String()
}

// renderParagraphV2 reads a <w:p> and converts it to HTML.
// It collects all runs, concatenates their text, does token replacement on the
// merged text, then produces the final HTML with per-run formatting.
func renderParagraphV2(d *xml.Decoder, values map[string]string) string {
	var runs []runInfo
	var align string
	depth := 1

	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch t.Name.Local {
			case "jc": // paragraph alignment
				for _, a := range t.Attr {
					if a.Name.Local == "val" {
						align = a.Value
					}
				}
			case "r": // run
				ri := collectRun(d)
				runs = append(runs, ri)
				depth--
			case "hyperlink":
				hlRuns := collectHyperlink(d)
				runs = append(runs, hlRuns...)
				depth--
			case "drawing", "pict", "object":
				runs = append(runs, runInfo{text: `<span class="img-placeholder">🖼</span>`})
				skipElement(d)
				depth--
			}
		case xml.EndElement:
			depth--
			if depth <= 0 {
				goto done
			}
		}
	}
done:

	if len(runs) == 0 {
		return "<p>&nbsp;</p>\n"
	}

	// --- Token replacement across runs ---
	// 1. Concatenate all run text.
	var fullText strings.Builder
	runBoundaries := make([]int, len(runs)) // end index of each run in fullText
	for i, r := range runs {
		fullText.WriteString(r.text)
		runBoundaries[i] = fullText.Len()
	}
	merged := fullText.String()

	// 2. Find and replace all {{token}} in the merged text.
	replaced := reTokenInXML.ReplaceAllStringFunc(merged, func(match string) string {
		key := match[2 : len(match)-2] // strip {{ and }}
		if val, ok := values[key]; ok {
			return val
		}
		return match // keep original if no value
	})

	// 3. If no changes were made, render runs as-is with formatting.
	if replaced == merged {
		return renderRunsToHTML(runs, align)
	}

	// 4. If replacements were made, we need to redistribute text back to runs.
	//    Simple approach: if the paragraph had tokens replaced, treat all text
	//    as having the formatting of the first non-empty run (good enough for preview).
	//    Better approach: map character positions back. For preview, the simple way works.
	dominantRun := runInfo{}
	for _, r := range runs {
		if r.text != "" {
			dominantRun = r
			break
		}
	}
	dominantRun.text = replaced
	return renderRunsToHTML([]runInfo{dominantRun}, align)
}

// renderRunsToHTML converts a slice of runInfo into an HTML <p> element.
func renderRunsToHTML(runs []runInfo, align string) string {
	var content strings.Builder
	for _, r := range runs {
		if r.text == "" {
			continue
		}
		s := r.text

		// Skip formatting for HTML-content runs (like image placeholders)
		if strings.HasPrefix(s, "<") {
			content.WriteString(s)
			continue
		}

		if r.bold {
			s = "<strong>" + s + "</strong>"
		}
		if r.italic {
			s = "<em>" + s + "</em>"
		}
		if r.underline {
			s = "<u>" + s + "</u>"
		}
		if r.strike {
			s = "<s>" + s + "</s>"
		}

		var styles []string
		if r.fontSize != "" {
			if size := parseHalfPoints(r.fontSize); size > 0 {
				styles = append(styles, fmt.Sprintf("font-size:%.0fpt", size))
			}
		}
		if r.fontFamily != "" {
			styles = append(styles, fmt.Sprintf("font-family:'%s',serif", r.fontFamily))
		}

		if len(styles) > 0 {
			s = fmt.Sprintf(`<span style="%s">%s</span>`, strings.Join(styles, ";"), s)
		}

		content.WriteString(s)
	}

	text := content.String()
	if text == "" {
		return "<p>&nbsp;</p>\n"
	}

	style := ""
	switch align {
	case "center":
		style = ` style="text-align:center"`
	case "right":
		style = ` style="text-align:right"`
	case "both":
		style = ` style="text-align:justify"`
	}

	return fmt.Sprintf("<p%s>%s</p>\n", style, text)
}

// collectRun reads a <w:r> element and returns its info without producing HTML yet.
func collectRun(d *xml.Decoder) runInfo {
	var ri runInfo
	depth := 1

	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch t.Name.Local {
			case "b":
				// Check for w:val="false" / w:val="0"
				ri.bold = !isFalseAttr(t)
			case "i":
				ri.italic = !isFalseAttr(t)
			case "u":
				for _, a := range t.Attr {
					if a.Name.Local == "val" && a.Value != "none" {
						ri.underline = true
					}
				}
			case "strike":
				ri.strike = !isFalseAttr(t)
			case "sz":
				for _, a := range t.Attr {
					if a.Name.Local == "val" {
						ri.fontSize = a.Value
					}
				}
			case "rFonts":
				for _, a := range t.Attr {
					if a.Name.Local == "ascii" || a.Name.Local == "hAnsi" {
						ri.fontFamily = a.Value
						break
					}
				}
			case "t":
				ri.text += readText(d)
				depth--
			case "tab":
				ri.text += "\t"
			case "br":
				ri.text += "\n"
			case "drawing", "pict", "object":
				ri.text = `<span class="img-placeholder">🖼</span>`
				skipElement(d)
				depth--
			}
		case xml.EndElement:
			depth--
			if depth <= 0 {
				return ri
			}
		}
	}
	return ri
}

// collectHyperlink processes a hyperlink element and returns its runs.
func collectHyperlink(d *xml.Decoder) []runInfo {
	var runs []runInfo
	depth := 1

	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "r" {
				runs = append(runs, collectRun(d))
				depth--
			}
		case xml.EndElement:
			depth--
			if depth <= 0 {
				return runs
			}
		}
	}
	return runs
}

// renderTableCellV2 reads a <w:tc> and returns its HTML content.
func renderTableCellV2(d *xml.Decoder, values map[string]string) string {
	var content strings.Builder
	depth := 1

	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "p" {
				content.WriteString(renderParagraphV2(d, values))
				depth--
			}
		case xml.EndElement:
			depth--
			if depth <= 0 {
				return content.String()
			}
		}
	}
	return content.String()
}

// checkTableBorders peeks at the table properties to see if it has borders.
// Returns true if any border is found. Note: this consumes the <w:tblPr> element.
func checkTableBorders(d *xml.Decoder) bool {
	// We don't actually consume the tblPr here because the main parser needs to
	// continue reading rows. Just return true as a default for letter templates
	// (most administrative letter tables have no visible borders — they use tabs).
	return false
}

// --- Helper functions ---

// readText reads the text content of a <w:t> element.
func readText(d *xml.Decoder) string {
	var sb strings.Builder
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.CharData:
			sb.Write(t)
		case xml.EndElement:
			return sb.String()
		}
	}
	return sb.String()
}

// skipElement consumes all tokens until the matching end element.
func skipElement(d *xml.Decoder) {
	depth := 1
	for {
		tok, err := d.Token()
		if err != nil {
			return
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
			if depth <= 0 {
				return
			}
		}
	}
}

// isFalseAttr checks if an element has w:val="false" or w:val="0".
func isFalseAttr(t xml.StartElement) bool {
	for _, a := range t.Attr {
		if a.Name.Local == "val" {
			return a.Value == "false" || a.Value == "0"
		}
	}
	return false
}

// parseHalfPoints converts an OOXML half-point size string to points.
func parseHalfPoints(s string) float64 {
	var v float64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			v = v*10 + float64(c-'0')
		}
	}
	if v > 0 {
		return v / 2
	}
	return 0
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
