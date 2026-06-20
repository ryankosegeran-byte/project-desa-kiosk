//go:build !windows

package print

import (
	"fmt"
	"os"
)

// DocxRenderer is a non-Windows stub. Faithful DOCX rendering requires Microsoft
// Word (COM automation), which is only available on the Windows kiosk machines.
// This keeps the kiosk module buildable on other platforms (e.g. for tooling).
type DocxRenderer struct {
	outputDir string
}

func NewDocxRenderer(outputDir string) *DocxRenderer {
	_ = os.MkdirAll(outputDir, 0o755)
	return &DocxRenderer{outputDir: outputDir}
}

func (d *DocxRenderer) Close() error { return nil }

func (d *DocxRenderer) Render(docxBytes []byte, values map[string]string) (string, error) {
	return "", fmt.Errorf("render DOCX via Word hanya didukung di Windows")
}
