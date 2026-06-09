package print

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type Printer struct {
	SumatraPath string // Path to SumatraPDF executable
}

func NewPrinter(sumatraPath string) *Printer {
	if sumatraPath == "" {
		sumatraPath = "SumatraPDF.exe" // default in PATH or local directory
	}
	return &Printer{SumatraPath: sumatraPath}
}

func (p *Printer) PrintPDF(pdfPath string) error {
	// Verify file exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return fmt.Errorf("PDF file does not exist: %s", pdfPath)
	}

	// We only run SumatraPDF on windows
	if runtime.GOOS != "windows" {
		// For non-windows development/testing, we just log and skip the print command
		fmt.Printf("[MOCK PRINT] Simulating printing for path: %s on %s\n", pdfPath, runtime.GOOS)
		return nil
	}

	// Windows print command: SumatraPDF.exe -print-to-default -silent <pdfPath>
	cmd := exec.Command(p.SumatraPath, "-print-to-default", "-silent", pdfPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute print command: %w, output: %s", err, string(output))
	}

	return nil
}
