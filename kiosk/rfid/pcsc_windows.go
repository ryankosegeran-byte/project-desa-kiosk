//go:build windows

package rfid

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows"
)

// Minimal pure-Go PC/SC (winscard.dll) binding. No CGO required.
//
// This drives PC/SC readers such as the ACR122U: it waits for a card to be
// presented, then issues the standard PC/SC "Get Data (UID)" APDU
// (FF CA 00 00 00) to retrieve the card UID, which works for both ISO14443
// Type A and Type B cards exposed through the reader.

var (
	modwinscard = windows.NewLazySystemDLL("winscard.dll")

	procEstablishContext = modwinscard.NewProc("SCardEstablishContext")
	procReleaseContext   = modwinscard.NewProc("SCardReleaseContext")
	procIsValidContext   = modwinscard.NewProc("SCardIsValidContext")
	procListReadersW     = modwinscard.NewProc("SCardListReadersW")
	procGetStatusChangeW = modwinscard.NewProc("SCardGetStatusChangeW")
	procConnectW         = modwinscard.NewProc("SCardConnectW")
	procDisconnect       = modwinscard.NewProc("SCardDisconnect")
	procTransmit         = modwinscard.NewProc("SCardTransmit")
	procCancel           = modwinscard.NewProc("SCardCancel")
)

const (
	scardScopeSystem = 2

	scardShareShared = 2
	scardShareDirect = 3
	scardProtocolT0  = 1
	scardProtocolT1  = 2
	scardProtocolTx  = scardProtocolT0 | scardProtocolT1
	scardLeaveCard   = 0

	scardStateUnaware = 0x0000
	scardStateChanged = 0x0002
	scardStateEmpty   = 0x0010
	scardStatePresent = 0x0020

	scardSuccess        = 0x00000000
	scardETimeout       = 0x8010000A
	scardECancelled     = 0x80100002
	scardENoReadersAvl  = 0x8010002E
	scardEUnknownReader = 0x80100009
	scardWRemovedCard   = 0x80100069
	scardERemovedCard   = 0x80100069

	infiniteTimeout = 0xFFFFFFFF
)

// scardIORequest mirrors the SCARD_IO_REQUEST struct.
type scardIORequest struct {
	Protocol  uint32
	PciLength uint32
}

// readerState mirrors SCARD_READERSTATEW.
type readerState struct {
	Reader       *uint16
	UserData     uintptr
	CurrentState uint32
	EventState   uint32
	Atr          [36]byte
	AtrLen       uint32
}

func scardError(code uintptr) error {
	c := uint32(code)
	switch c {
	case scardETimeout:
		return errTimeout
	case scardECancelled:
		return context.Canceled
	case scardENoReadersAvl:
		return errNoReaders
	default:
		return fmt.Errorf("winscard error 0x%08X", c)
	}
}

var (
	errTimeout   = errors.New("winscard: timeout")
	errNoReaders = errors.New("winscard: tidak ada reader tersedia")
)

func establishContext() (uintptr, error) {
	var ctx uintptr
	r, _, _ := procEstablishContext.Call(
		uintptr(scardScopeSystem),
		0, 0,
		uintptr(unsafe.Pointer(&ctx)),
	)
	if r != scardSuccess {
		return 0, scardError(r)
	}
	return ctx, nil
}

func releaseContext(ctx uintptr) {
	procReleaseContext.Call(ctx)
}

// listReaders returns the names of all connected PC/SC readers.
func listReaders(ctx uintptr) ([]string, error) {
	var bufLen uint32
	r, _, _ := procListReadersW.Call(
		ctx, 0, 0,
		uintptr(unsafe.Pointer(&bufLen)),
	)
	if r != scardSuccess {
		return nil, scardError(r)
	}
	if bufLen == 0 {
		return nil, errNoReaders
	}

	buf := make([]uint16, bufLen)
	r, _, _ = procListReadersW.Call(
		ctx, 0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufLen)),
	)
	if r != scardSuccess {
		return nil, scardError(r)
	}
	return parseMultiString(buf), nil
}

// parseMultiString splits a Windows double-null-terminated multi-string.
func parseMultiString(buf []uint16) []string {
	var out []string
	start := 0
	for i := 0; i < len(buf); i++ {
		if buf[i] == 0 {
			if i == start {
				break // double null => end
			}
			out = append(out, syscall.UTF16ToString(buf[start:i]))
			start = i + 1
		}
	}
	return out
}

// waitForCard blocks until a card is present on the named reader, ctx is
// cancelled, or the timeout elapses. Returns true if a card is present.
func waitForCard(scardCtx uintptr, readerName string, currentState uint32) (uint32, bool, error) {
	namePtr, err := windows.UTF16PtrFromString(readerName)
	if err != nil {
		return currentState, false, err
	}

	rs := []readerState{{
		Reader:       namePtr,
		CurrentState: currentState,
	}}

	r, _, _ := procGetStatusChangeW.Call(
		scardCtx,
		uintptr(1000), // 1s timeout so we can re-check ctx cancellation
		uintptr(unsafe.Pointer(&rs[0])),
		uintptr(1),
	)
	if uint32(r) == scardETimeout {
		return rs[0].EventState, rs[0].EventState&scardStatePresent != 0, errTimeout
	}
	if r != scardSuccess {
		return currentState, false, scardError(r)
	}

	newState := rs[0].EventState &^ scardStateChanged
	present := newState&scardStatePresent != 0
	return newState, present, nil
}

// readCardUID connects to the card and returns its UID bytes.
func readCardUID(scardCtx uintptr, readerName string) ([]byte, error) {
	namePtr, err := windows.UTF16PtrFromString(readerName)
	if err != nil {
		return nil, err
	}

	var card uintptr
	var activeProtocol uint32
	r, _, _ := procConnectW.Call(
		scardCtx,
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(scardShareShared),
		uintptr(scardProtocolTx),
		uintptr(unsafe.Pointer(&card)),
		uintptr(unsafe.Pointer(&activeProtocol)),
	)
	if r != scardSuccess {
		return nil, scardError(r)
	}
	defer procDisconnect.Call(card, uintptr(scardLeaveCard))

	// Standard PC/SC pseudo-APDU to get the card UID.
	apdu := []byte{0xFF, 0xCA, 0x00, 0x00, 0x00}

	sendPci := scardIORequest{Protocol: activeProtocol, PciLength: 8}
	recv := make([]byte, 258)
	recvLen := uint32(len(recv))

	r, _, _ = procTransmit.Call(
		card,
		uintptr(unsafe.Pointer(&sendPci)),
		uintptr(unsafe.Pointer(&apdu[0])),
		uintptr(len(apdu)),
		0,
		uintptr(unsafe.Pointer(&recv[0])),
		uintptr(unsafe.Pointer(&recvLen)),
	)
	if r != scardSuccess {
		return nil, scardError(r)
	}
	if recvLen < 2 {
		return nil, errors.New("respon APDU terlalu pendek")
	}

	// Last two bytes are SW1 SW2; expect 90 00 on success.
	data := recv[:recvLen]
	sw1, sw2 := data[recvLen-2], data[recvLen-1]
	if sw1 != 0x90 || sw2 != 0x00 {
		return nil, fmt.Errorf("kartu menolak permintaan UID (SW=%02X%02X)", sw1, sw2)
	}
	return data[:recvLen-2], nil
}

// pickReader selects a reader name based on the optional filter.
func pickReader(readers []string, filter string) (string, bool) {
	if len(readers) == 0 {
		return "", false
	}
	if filter == "" {
		return readers[0], true
	}
	lf := strings.ToLower(filter)
	for _, r := range readers {
		if strings.Contains(strings.ToLower(r), lf) {
			return r, true
		}
	}
	return "", false
}

// runReaderLoop is the Windows implementation of the physical reader loop.
func runReaderLoop(ctx context.Context, broker *Broker, cfg ReaderConfig) error {
	if err := modwinscard.Load(); err != nil {
		return fmt.Errorf("winscard.dll tidak tersedia (pastikan layanan Smart Card aktif): %w", err)
	}

	var (
		lastUID string
		lastAt  time.Time
	)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		scardCtx, err := establishContext()
		if err != nil {
			log.Warn().Err(err).Msg("Gagal membuka konteks PC/SC, mencoba lagi...")
			if !sleepCtx(ctx, cfg.PollInterval) {
				return nil
			}
			continue
		}

		readerName, err := awaitReader(ctx, scardCtx, cfg)
		if err != nil {
			releaseContext(scardCtx)
			if ctx.Err() != nil {
				return nil
			}
			if !sleepCtx(ctx, cfg.PollInterval) {
				return nil
			}
			continue
		}

		log.Info().Str("reader", readerName).Msg("Pembaca RFID PC/SC siap")
		pollReader(ctx, scardCtx, readerName, broker, cfg, &lastUID, &lastAt)
		releaseContext(scardCtx)

		if ctx.Err() != nil {
			return nil
		}
	}
}

// awaitReader waits until a matching reader is connected.
func awaitReader(ctx context.Context, scardCtx uintptr, cfg ReaderConfig) (string, error) {
	for {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		readers, err := listReaders(scardCtx)
		if err == nil {
			if name, ok := pickReader(readers, cfg.ReaderNameFilter); ok {
				return name, nil
			}
		}
		if !sleepCtx(ctx, cfg.PollInterval) {
			return "", ctx.Err()
		}
	}
}

// pollReader watches a single reader for card insertions until it disconnects
// or ctx is cancelled.
func pollReader(ctx context.Context, scardCtx uintptr, readerName string, broker *Broker, cfg ReaderConfig, lastUID *string, lastAt *time.Time) {
	state := uint32(scardStateUnaware)
	for {
		if ctx.Err() != nil {
			return
		}

		newState, present, err := waitForCard(scardCtx, readerName, state)
		if err != nil && !errors.Is(err, errTimeout) {
			// Reader likely removed or error; bail to re-scan readers.
			log.Warn().Err(err).Str("reader", readerName).Msg("Pembaca RFID terputus / error")
			return
		}
		state = newState

		if present {
			uid, rerr := readCardUID(scardCtx, readerName)
			if rerr != nil {
				log.Debug().Err(rerr).Msg("Gagal membaca UID kartu")
			} else {
				publishUID(broker, uid, cfg.UIDFormat, lastUID, lastAt)
			}
		}
	}
}

// sleepCtx sleeps for d or until ctx is cancelled. Returns false if cancelled.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
