package rfid

import "testing"

func TestFormatUID(t *testing.T) {
	raw := []byte{0x04, 0xA1, 0xB2, 0xC3}

	cases := []struct {
		format string
		want   string
	}{
		{"", "04A1B2C3"},
		{"hex", "04A1B2C3"},
		{"HEX", "04A1B2C3"},
		{"hex-colon", "04:A1:B2:C3"},
		{"decimal", "3283263748"},
	}

	for _, c := range cases {
		if got := formatUID(raw, c.format); got != c.want {
			t.Errorf("formatUID(%q) = %q, want %q", c.format, got, c.want)
		}
	}
}

func TestFormatUIDEmpty(t *testing.T) {
	if got := formatUID(nil, "hex"); got != "" {
		t.Errorf("formatUID(nil) = %q, want empty", got)
	}
}

func TestDecimalLE(t *testing.T) {
	if got := decimalLE([]byte{0x01, 0x00, 0x00, 0x00}); got != "1" {
		t.Errorf("decimalLE little-endian = %q, want 1", got)
	}
	if got := decimalLE([]byte{0x00}); got != "0" {
		t.Errorf("decimalLE zero = %q, want 0", got)
	}
}
