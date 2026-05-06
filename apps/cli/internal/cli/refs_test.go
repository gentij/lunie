package cli

import "testing"

func TestParsePositiveIntArg(t *testing.T) {
	value, err := parsePositiveIntArg("run number", "42")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != 42 {
		t.Fatalf("expected 42, got %d", value)
	}
}

func TestParsePositiveIntArgRejectsInvalidValues(t *testing.T) {
	for _, raw := range []string{"0", "-1", "abc", ""} {
		if _, err := parsePositiveIntArg("run number", raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}
