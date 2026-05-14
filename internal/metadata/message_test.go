package metadata

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestRecord_Encode(t *testing.T) {
	r := &Record{
		Offset:    10,
		Timestamp: 100,
		Key:       []byte("key"),
		Value:     []byte("value"),
	}

	encoded := r.Encode()

	// Expected length: 8 (Offset) + 8 (Timestamp) + 4 (KeyLen) + 3 (Key) + 4 (ValueLen) + 5 (Value) = 32
	expectedLen := 32
	if len(encoded) != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, len(encoded))
	}

	// Verify content
	if len(encoded) >= 8 {
		offset := binary.BigEndian.Uint64(encoded[0:8])
		if offset != r.Offset {
			t.Errorf("Expected offset %d, got %d", r.Offset, offset)
		}
	}

	// Key should start at index 20 (8+8+4)
	if !bytes.Contains(encoded, []byte("key")) {
		t.Error("Encoded record does not contain key")
	}
	// Value should start at index 27 (8+8+4+3+4)
	if !bytes.Contains(encoded, []byte("value")) {
		t.Error("Encoded record does not contain value")
	}
}
