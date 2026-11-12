package pkt

import (
	"bytes"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedData  [][]byte
		expectedFlush []bool
		expectError   bool
	}{
		{
			name:          "standard packets with flush",
			input:         "0006a\n0009hello0000",
			expectedData:  [][]byte{[]byte("a\n"), []byte("hello"), nil},
			expectedFlush: []bool{false, false, true},
			expectError:   false,
		},
		{
			name:          "standard packets with flush in between",
			input:         "0006a\n00000006b\n",
			expectedData:  [][]byte{[]byte("a\n"), nil, []byte("b\n")},
			expectedFlush: []bool{false, true, false},
			expectError:   false,
		},
		{
			name:          "only flush packet",
			input:         "0000",
			expectedData:  [][]byte{nil},
			expectedFlush: []bool{true},
			expectError:   false,
		},
		{
			name:        "malformed hex length",
			input:       "ZZZZsome-data",
			expectError: true,
		},
		{
			name:        "unexpected EOF mid-packet",
			input:       "000aabc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packets, err := Parse(strings.NewReader(tt.input))

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(packets) != len(tt.expectedData) {
				t.Fatalf("expected %d packets, got %d", len(tt.expectedData), len(packets))
			}

			for i, packet := range packets {
				flush := packet.Len() == 0
				if flush != tt.expectedFlush[i] {
					t.Errorf("packet %d: expected flush=%v, got flush=%v", i, tt.expectedFlush[i], flush)
				}
				if !bytes.Equal(packet.Data(), tt.expectedData[i]) {
					t.Errorf("packet %d: expected data %q, got %q", i, tt.expectedData[i], packet.Data())
				}
			}

		})
	}
}
