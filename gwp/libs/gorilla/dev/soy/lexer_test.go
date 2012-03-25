package soy

import (
	"testing"
)

func TestScanNumber(t *testing.T) {
	validIntegers := []string{
		// Decimal.
		"42",
		"-827",
		// Hexadecimal.
		"0x1A2B",
	}
	invalidIntegers := []string{
		// Decimal.
		"042",
		"-0827",
		// Hexadecimal.
		"-0x1A2B",
		"0X1A2B",
		"0x1a2b",
		"0x1A2B.2B",
	}
	validFloats := []string{
		"0.5",
		"-100.0",
		"-3e-3",
		"6.02e23",
		"5.1e-9",
	}
	invalidFloats := []string{
		".5",
		"-.5",
		"100.",
		"-100.",
		"-3E-3",
		"6.02E23",
		"5.1E-9",
		"-3e",
		"6.02e",
	}

	for _, v := range validIntegers {
		l := newLexer("", v)
		typ, ok := scanNumber(l)
		res := l.input[l.start:l.pos]
		if !ok || typ != tokenInteger {
			t.Fatalf("Expected a valid integer for %q", v)
		}
		if res != v {
			t.Fatalf("Expected %q, got %q", v, res)
		}
	}
	for _, v := range invalidIntegers {
		l := newLexer("", v)
		_, ok := scanNumber(l)
		if ok {
			t.Fatalf("Expected an invalid integer for %q", v)
		}
	}
	for _, v := range validFloats {
		l := newLexer("", v)
		typ, ok := scanNumber(l)
		res := l.input[l.start:l.pos]
		if !ok || typ != tokenFloat {
			t.Fatalf("Expected a valid float for %q", v)
		}
		if res != v {
			t.Fatalf("Expected %q, got %q", v, res)
		}
	}
	for _, v := range invalidFloats {
		l := newLexer("", v)
		_, ok := scanNumber(l)
		if ok {
			t.Fatalf("Expected an invalid float for %q", v)
		}
	}
}
