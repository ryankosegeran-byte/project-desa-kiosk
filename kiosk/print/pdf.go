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

// GeneratePDF generates a PDF file from HTML template and data using chromedp
func (g *PDFGenerator) GeneratePDF(ctx context.Context, templateHTML string, warga *models.Warga, dataSurat map[string]interface{}, dateToday string) (string, error) {
	// 1. Parse and execute template
	tmpl, err := template.New("surat").Parse(templateHTML)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	tmplData := map[string]interface{}{
		"Warga":     warga,
		"DataSurat": dataSurat,
		"DateToday": dateToday,
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

	// 4. Generate PDF using printToPDF
	var pdfBuffer []byte
	err = chromedp.Run(chromeCtx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPaperWidth(8.27).   // A4 Width in inches
				WithPaperHeight(11.69).  // A4 Height in inches
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
