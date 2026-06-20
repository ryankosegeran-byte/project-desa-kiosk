//go:build windows

package docx

import (
	"fmt"
	"os"
	"runtime"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const (
	wdFindContinue    = 1
	wdReplaceAll      = 2
	wdExportFormatPDF = 17
)

// RenderToPDF fills {{token}} placeholders in a DOCX template and returns PDF bytes,
// rendered by Microsoft Word via COM automation.
// Word's Find/Replace works on logical text, so split-run tokens are handled correctly.
func RenderToPDF(docxBytes []byte, values map[string]string) ([]byte, error) {
	if len(docxBytes) == 0 {
		return nil, fmt.Errorf("template DOCX kosong")
	}

	type result struct {
		pdf []byte
		err error
	}
	ch := make(chan result, 1)

	// COM apartments are thread-affine: must create and use Word on the same OS thread.
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if err := ole.CoInitialize(0); err != nil {
			ch <- result{err: fmt.Errorf("CoInitialize gagal: %w", err)}
			return
		}
		defer ole.CoUninitialize()

		tmpDocx, err := os.CreateTemp("", "render-*.docx")
		if err != nil {
			ch <- result{err: fmt.Errorf("gagal buat temp docx: %w", err)}
			return
		}
		tmpDocxPath := tmpDocx.Name()
		if _, werr := tmpDocx.Write(docxBytes); werr != nil {
			tmpDocx.Close()
			os.Remove(tmpDocxPath)
			ch <- result{err: werr}
			return
		}
		tmpDocx.Close()
		defer os.Remove(tmpDocxPath)

		tmpPDF, err := os.CreateTemp("", "render-*.pdf")
		if err != nil {
			ch <- result{err: fmt.Errorf("gagal buat temp pdf: %w", err)}
			return
		}
		tmpPDFPath := tmpPDF.Name()
		tmpPDF.Close()
		defer os.Remove(tmpPDFPath)

		unknown, err := oleutil.CreateObject("Word.Application")
		if err != nil {
			ch <- result{err: fmt.Errorf("gagal memulai Microsoft Word (pastikan terpasang): %w", err)}
			return
		}
		word, err := unknown.QueryInterface(ole.IID_IDispatch)
		unknown.Release()
		if err != nil {
			ch <- result{err: fmt.Errorf("QueryInterface Word gagal: %w", err)}
			return
		}
		defer func() {
			oleutil.CallMethod(word, "Quit")
			word.Release()
		}()
		oleutil.PutProperty(word, "Visible", false)
		oleutil.PutProperty(word, "DisplayAlerts", 0)
		oleutil.PutProperty(word, "AutomationSecurity", 3)

		documents := oleutil.MustGetProperty(word, "Documents").ToIDispatch()
		defer documents.Release()

		docV, err := oleutil.CallMethod(documents, "Open", tmpDocxPath)
		if err != nil {
			ch <- result{err: fmt.Errorf("gagal membuka docx di Word: %w", err)}
			return
		}
		doc := docV.ToIDispatch()
		defer func() {
			oleutil.CallMethod(doc, "Close", false)
			doc.Release()
		}()

		for key, val := range values {
			if rerr := wordFindReplace(doc, "{{"+key+"}}", val); rerr != nil {
				ch <- result{err: fmt.Errorf("gagal mengganti token %q: %w", key, rerr)}
				return
			}
		}

		if _, eerr := oleutil.CallMethod(doc, "ExportAsFixedFormat", tmpPDFPath, wdExportFormatPDF); eerr != nil {
			ch <- result{err: fmt.Errorf("gagal export PDF: %w", eerr)}
			return
		}

		pdf, err := os.ReadFile(tmpPDFPath)
		if err != nil {
			ch <- result{err: fmt.Errorf("gagal baca PDF output: %w", err)}
			return
		}
		ch <- result{pdf: pdf}
	}()

	res := <-ch
	return res.pdf, res.err
}

func wordFindReplace(doc *ole.IDispatch, find, repl string) error {
	content := oleutil.MustGetProperty(doc, "Content").ToIDispatch()
	defer content.Release()
	findObj := oleutil.MustGetProperty(content, "Find").ToIDispatch()
	defer findObj.Release()
	_, err := oleutil.CallMethod(findObj, "Execute",
		find, true, false, false, false, false, true, wdFindContinue, false, repl, wdReplaceAll,
	)
	return err
}
