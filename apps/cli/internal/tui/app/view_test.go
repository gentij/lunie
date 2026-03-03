package app

import "testing"

func TestPadPlainLines_PadsAndTruncatesToViewport(t *testing.T) {
	lines := padPlainLines("abc\n12345\nxyz", 4, 4)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}
	if lines[0] != "abc " {
		t.Fatalf("unexpected first line: %q", lines[0])
	}
	if lines[1] != "1234" {
		t.Fatalf("unexpected second line: %q", lines[1])
	}
	if lines[2] != "xyz " {
		t.Fatalf("unexpected third line: %q", lines[2])
	}
	if lines[3] != "    " {
		t.Fatalf("unexpected padded line: %q", lines[3])
	}
}

func TestPlainSlice_HandlesBounds(t *testing.T) {
	line := "abcdef"
	if got := plainSlice(line, 1, 4); got != "bcd" {
		t.Fatalf("unexpected slice result: %q", got)
	}
	if got := plainSlice(line, -4, 2); got != "ab" {
		t.Fatalf("unexpected negative-start result: %q", got)
	}
	if got := plainSlice(line, 4, 99); got != "ef" {
		t.Fatalf("unexpected wide-end result: %q", got)
	}
	if got := plainSlice(line, 5, 5); got != "" {
		t.Fatalf("expected empty result for equal bounds, got %q", got)
	}
}
