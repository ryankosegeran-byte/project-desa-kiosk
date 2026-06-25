//go:build !windows

package rfid

import (
	"context"
	"errors"
)

// runReaderLoop is a no-op on non-Windows platforms. The physical PC/SC reader
// integration currently targets Windows (winscard.dll). On other platforms the
// kiosk still works via mock scans and keyboard-wedge readers.
func runReaderLoop(_ context.Context, _ *Broker, _ ReaderConfig) error {
	return errors.New("pembaca RFID PC/SC hanya didukung di Windows pada build ini")
}
