//go:build !windows

package docx

import "fmt"

// RenderToPDF is not supported on this platform without LibreOffice.
// For FreeBSD/Linux servers, install LibreOffice and implement soffice --headless here.
func RenderToPDF(_ []byte, _ map[string]string) ([]byte, error) {
	return nil, fmt.Errorf("render PDF dari DOCX memerlukan Microsoft Word (Windows) atau LibreOffice")
}
