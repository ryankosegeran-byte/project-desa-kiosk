package print

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/project-desa-kiosk/internal/models"
)

// Paper format dimensions in inches
const (
	// A4: 210mm x 297mm = 8.27" x 11.69"
	A4Width  = 8.27
	A4Height = 11.69

	// F4 / Folio: 215mm x 330mm = 8.46" x 12.99"
	F4Width  = 8.46
	F4Height = 12.99
)

// FormatKertas constants
const (
	FormatKertasA4 = "A4"
	FormatKertasF4 = "F4"
)

type PDFGenerator struct {
	OutputDir string
}

func NewPDFGenerator(outputDir string) *PDFGenerator {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		// Log error to stdout for debugging
		fmt.Printf("Warning: failed to create output directory %s: %v\n", outputDir, err)
	}
	return &PDFGenerator{OutputDir: outputDir}
}

// FormatIndonesianDate formats a time.Time to Indonesian format (e.g. "6 Juni 2026")
func FormatIndonesianDate(t time.Time) string {
	months := map[time.Month]string{
		time.January:   "Januari",
		time.February:  "Februari",
		time.March:     "Maret",
		time.April:     "April",
		time.May:       "Mei",
		time.June:      "Juni",
		time.July:      "Juli",
		time.August:    "Agustus",
		time.September: "September",
		time.October:   "Oktober",
		time.November:  "November",
		time.December:  "Desember",
	}
	return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()], t.Year())
}

// RenderHTML executes the HTML template with the same data shape used for
// printing and returns the final HTML string. The kiosk UI uses this for an
// accurate, server-rendered preview (identical to the printed output) instead
// of doing fragile client-side token replacement.
func RenderHTML(templateHTML string, warga *models.Warga, dataSurat map[string]interface{}, dateToday string, nomorSurat string, desaKepalaDesa string, desaNIP string) (string, error) {
	tmpl, err := template.New("surat").Parse(templateHTML)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	tmplData := map[string]interface{}{
		"Warga":          warga,
		"DataSurat":      dataSurat,
		"DateToday":      dateToday,
		"NomorSurat":     nomorSurat,
		"DesaKepalaDesa": desaKepalaDesa,
		"DesaNIP":        desaNIP,
	}
	if err := tmpl.Execute(&buf, tmplData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}
// getPaperDimensions returns width and height in inches based on format
func getPaperDimensions(formatKertas string) (width, height float64) {
	switch formatKertas {
	case FormatKertasF4:
		return F4Width, F4Height
	default: // A4 or default
		return A4Width, A4Height
	}
}

// GeneratePDF generates a PDF file from HTML template and data using chromedp
// Supports both A4 and F4 paper formats based on formatKertas parameter
func (g *PDFGenerator) GeneratePDF(ctx context.Context, templateHTML string, warga *models.Warga, dataSurat map[string]interface{}, dateToday string, nomorSurat string, desaKepalaDesa string, desaNIP string, formatKertas string) (string, error) {
	// Get paper dimensions
	paperWidth, paperHeight := getPaperDimensions(formatKertas)

	// 1. Parse and execute template
	tmpl, err := template.New("surat").Parse(templateHTML)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	tmplData := map[string]interface{}{
		"Warga":          warga,
		"DataSurat":      dataSurat,
		"DateToday":      dateToday,
		"NomorSurat":     nomorSurat,
		"DesaKepalaDesa": desaKepalaDesa,
		"DesaNIP":        desaNIP,
	}

	if err := tmpl.Execute(&buf, tmplData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// 2. Write executed HTML to a temporary file
	tmpFile, err := os.CreateTemp("", "surat-*.html")
	if err != nil {
		return "", fmt.Errorf("failed to create temp HTML file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(buf.Bytes()); err != nil {
		return "", fmt.Errorf("failed to write temp HTML: %w", err)
	}

	// Get absolute path of temp file for chromedp
	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	url := "file://" + filepath.ToSlash(absPath)

	// 3. Setup chromedp context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	chromeCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 4. Generate PDF using printToPDF with dynamic paper size
	var pdfBuffer []byte
	err = chromedp.Run(chromeCtx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPaperWidth(paperWidth).
				WithPaperHeight(paperHeight).
				WithMarginTop(0.4).
				WithMarginBottom(0.4).
				WithMarginLeft(0.4).
				WithMarginRight(0.4).
				WithPrintBackground(true).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBuffer = buf
			return nil
		}),
	)
	if err != nil {
		return "", fmt.Errorf("chromedp run failed: %w", err)
	}

	// Ensure output directory exists before writing
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to ensure output directory: %w", err)
	}

	// 5. Write PDF to destination path
	destFile := filepath.Join(g.OutputDir, fmt.Sprintf("surat-%s.pdf", uuidString()))
	if err := os.WriteFile(destFile, pdfBuffer, 0644); err != nil {
		return "", fmt.Errorf("failed to write pdf file: %w", err)
	}

	return destFile, nil
}

// Simple helper to avoid dependency cycles if we need a UUID
func uuidString() string {
	now := time.Now().UnixNano()
	return fmt.Sprintf("%d", now)
}
