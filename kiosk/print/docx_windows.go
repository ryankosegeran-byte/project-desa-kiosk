//go:build windows

package print

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// Word automation constants.
const (
	wdFindContinue    = 1
	wdReplaceAll      = 2
	wdExportFormatPDF = 17
)

// DocxRenderer fills DOCX templates ({{token}} placeholders) and exports them to
// PDF using Microsoft Word via COM automation.
//
// COM apartments are thread-affine: a Word instance must be created and used on
// one OS thread. To keep Word "warm" across requests (fast live previews), all
// rendering runs on a single dedicated, OS-thread-locked worker goroutine; Render
// hands jobs to it over a channel. The first render cold-starts Word (~10s),
// subsequent renders reuse it (~1-2s).
type DocxRenderer struct {
	outputDir string
	reqCh     chan docxJob
	quitCh    chan struct{}
	doneCh    chan struct{}
	closeOnce sync.Once
}

type docxJob struct {
	docxBytes []byte
	values    map[string]string
	respCh    chan docxResult
}

type docxResult struct {
	path string
	err  error
}

// NewDocxRenderer starts the render worker. Word itself is created lazily on the
// first Render so kiosk startup stays fast.
func NewDocxRenderer(outputDir string) *DocxRenderer {
	_ = os.MkdirAll(outputDir, 0o755)
	d := &DocxRenderer{
		outputDir: outputDir,
		reqCh:     make(chan docxJob),
		quitCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
	go d.worker()
	return d
}

// Close stops the worker and quits the warm Word instance. Safe to call once.
func (d *DocxRenderer) Close() error {
	d.closeOnce.Do(func() {
		close(d.quitCh)
		<-d.doneCh
	})
	return nil
}

// Render fills the template DOCX with values (token key -> value) and exports a
// PDF, returning its path. Safe for concurrent callers (serialized on the worker).
func (d *DocxRenderer) Render(docxBytes []byte, values map[string]string) (string, error) {
	if len(docxBytes) == 0 {
		return "", fmt.Errorf("template DOCX kosong")
	}
	respCh := make(chan docxResult, 1)
	select {
	case d.reqCh <- docxJob{docxBytes: docxBytes, values: values, respCh: respCh}:
	case <-d.doneCh:
		return "", fmt.Errorf("renderer DOCX sudah ditutup")
	}
	res := <-respCh
	return res.path, res.err
}

// worker owns the COM apartment and the warm Word instance for its entire life.
func (d *DocxRenderer) worker() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(d.doneCh)

	_ = ole.CoInitialize(0)
	defer ole.CoUninitialize()

	var word *ole.IDispatch
	quitWord := func() {
		if word == nil {
			return
		}
		func() {
			defer func() { _ = recover() }()
			oleutil.CallMethod(word, "Quit")
			word.Release()
		}()
		word = nil
	}
	defer quitWord()

	ensureWord := func() (*ole.IDispatch, error) {
		if word != nil {
			return word, nil
		}
		unknown, err := oleutil.CreateObject("Word.Application")
		if err != nil {
			return nil, fmt.Errorf("gagal memulai Microsoft Word (pastikan terpasang): %w", err)
		}
		w, err := unknown.QueryInterface(ole.IID_IDispatch)
		unknown.Release()
		if err != nil {
			return nil, fmt.Errorf("QueryInterface Word gagal: %w", err)
		}
		oleutil.PutProperty(w, "Visible", false)
		oleutil.PutProperty(w, "DisplayAlerts", 0)
		oleutil.PutProperty(w, "AutomationSecurity", 3) // force-disable macros
		word = w
		return word, nil
	}

	for {
		select {
		case <-d.quitCh:
			return
		case job := <-d.reqCh:
			path, err := d.runJob(ensureWord, job)
			if err != nil {
				// Word may be wedged after an error; drop it so the next job
				// cold-starts a clean instance.
				quitWord()
			}
			job.respCh <- docxResult{path: path, err: err}
		}
	}
}

// runJob renders a single document on the warm Word instance. Panics from the COM
// layer are converted into errors so one bad template cannot crash the kiosk.
func (d *DocxRenderer) runJob(ensureWord func() (*ole.IDispatch, error), job docxJob) (path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("render Word panik: %v", r)
		}
	}()

	word, err := ensureWord()
	if err != nil {
		return "", err
	}

	// Write the template to a temp .docx in outputDir (same volume as output).
	tmp, err := os.CreateTemp(d.outputDir, "render-*.docx")
	if err != nil {
		return "", fmt.Errorf("gagal buat temp docx: %w", err)
	}
	tmpPath := tmp.Name()
	if _, werr := tmp.Write(job.docxBytes); werr != nil {
		tmp.Close()
		return "", fmt.Errorf("gagal tulis temp docx: %w", werr)
	}
	tmp.Close()
	defer os.Remove(tmpPath)

	// Word COM resolves relative paths against its own working directory
	// (typically C:\Windows\System32), so always hand it absolute paths.
	tmpAbs, err := filepath.Abs(tmpPath)
	if err != nil {
		return "", fmt.Errorf("gagal resolve path docx: %w", err)
	}
	tmpPath = tmpAbs

	outPath := filepath.Join(d.outputDir, fmt.Sprintf("surat-%d.pdf", time.Now().UnixNano()))
	if outAbs, aerr := filepath.Abs(outPath); aerr == nil {
		outPath = outAbs
	}

	documents := oleutil.MustGetProperty(word, "Documents").ToIDispatch()
	defer documents.Release()

	docV, err := oleutil.CallMethod(documents, "Open", tmpPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuka docx di Word: %w", err)
	}
	doc := docV.ToIDispatch()
	defer func() {
		oleutil.CallMethod(doc, "Close", false) // wdDoNotSaveChanges
		doc.Release()
	}()

	for key, val := range job.values {
		if rerr := wordReplaceAll(doc, "{{"+key+"}}", val); rerr != nil {
			return "", fmt.Errorf("gagal mengganti token %q: %w", key, rerr)
		}
	}

	if _, eerr := oleutil.CallMethod(doc, "ExportAsFixedFormat", outPath, wdExportFormatPDF); eerr != nil {
		return "", fmt.Errorf("gagal export PDF: %w", eerr)
	}

	return outPath, nil
}

// wordReplaceAll replaces all occurrences of find with repl across the document
// body using Word's Find/Replace, which operates on logical text (split runs OK).
//
// Note: Word's ReplaceWith has a ~255 character limit; letter fields (names,
// addresses) stay well under that.
func wordReplaceAll(doc *ole.IDispatch, find, repl string) error {
	content := oleutil.MustGetProperty(doc, "Content").ToIDispatch()
	defer content.Release()
	findObj := oleutil.MustGetProperty(content, "Find").ToIDispatch()
	defer findObj.Release()

	// Find.Execute(FindText, MatchCase, MatchWholeWord, MatchWildcards,
	//   MatchSoundsLike, MatchAllWordForms, Forward, Wrap, Format, ReplaceWith, Replace)
	_, err := oleutil.CallMethod(findObj, "Execute",
		find, true, false, false, false, false, true, wdFindContinue, false, repl, wdReplaceAll,
	)
	return err
}
