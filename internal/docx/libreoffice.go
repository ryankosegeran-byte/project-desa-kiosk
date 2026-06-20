package docx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ConvertDocxToPDF converts DOCX bytes to PDF using LibreOffice headless mode.
// Each call uses an isolated user profile so concurrent requests don't conflict.
// Returns an error if LibreOffice is not installed.
func ConvertDocxToPDF(docxBytes []byte) ([]byte, error) {
	sofficePath, err := findSoffice()
	if err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "lo-preview-*")
	if err != nil {
		return nil, fmt.Errorf("gagal buat temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	docxPath := filepath.Join(tmpDir, "input.docx")
	if err := os.WriteFile(docxPath, docxBytes, 0600); err != nil {
		return nil, fmt.Errorf("gagal tulis DOCX temp: %w", err)
	}

	// Isolated profile URI so multiple concurrent calls don't conflict.
	// Windows needs file:///C:/path (3 slashes); Unix path already starts with /
	// so file:// + /path = file:///path (also 3 slashes).
	profileSlash := filepath.ToSlash(filepath.Join(tmpDir, "lo-profile"))
	var profileURI string
	if runtime.GOOS == "windows" {
		profileURI = "file:///" + profileSlash
	} else {
		profileURI = "file://" + profileSlash
	}

	cmd := exec.Command(sofficePath,
		"-env:UserInstallation="+profileURI,
		"--headless",
		"--norestore",
		"--nofirststartwizard",
		"--convert-to", "pdf:writer_pdf_Export",
		"--outdir", tmpDir,
		docxPath,
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("LibreOffice error: %w\n%s", err, string(out))
	}

	pdfPath := filepath.Join(tmpDir, "input.pdf")
	pdf, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("PDF output tidak ditemukan (LibreOffice mungkin gagal diam-diam): %w", err)
	}
	return pdf, nil
}

func findSoffice() (string, error) {
	if p, err := exec.LookPath("soffice"); err == nil {
		return p, nil
	}

	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			`C:\Program Files\LibreOffice\program\soffice.exe`,
			`C:\Program Files (x86)\LibreOffice\program\soffice.exe`,
		}
	} else {
		candidates = []string{
			"/usr/local/bin/soffice",
			"/usr/bin/soffice",
			"/opt/libreoffice/program/soffice",
			"/opt/libreoffice7.6/program/soffice",
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("LibreOffice tidak ditemukan — install dari https://www.libreoffice.org/download/")
	}
	return "", fmt.Errorf("LibreOffice tidak ditemukan — install: pkg install libreoffice (FreeBSD) / apt install libreoffice (Debian)")
}
