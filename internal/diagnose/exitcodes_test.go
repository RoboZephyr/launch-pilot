package diagnose

import (
	"strings"
	"testing"
)

func TestExplainExitCode_KnownCodes(t *testing.T) {
	tests := []struct {
		code            int
		wantMeaning     string
		wantSuggestion  string
		emptySuggestion bool
	}{
		{0, "Normal exit", "", true},
		{1, "General error", "Check program logs", false},
		{78, "Configuration error", "plutil -lint", false},
	}
	for _, tt := range tests {
		t.Run(tt.wantMeaning, func(t *testing.T) {
			meaning, suggestion := ExplainExitCode(tt.code)
			if !strings.Contains(meaning, tt.wantMeaning) {
				t.Errorf("ExplainExitCode(%d) meaning = %q, want it to contain %q", tt.code, meaning, tt.wantMeaning)
			}
			if tt.emptySuggestion {
				if suggestion != "" {
					t.Errorf("ExplainExitCode(%d) suggestion = %q, want empty", tt.code, suggestion)
				}
			} else {
				if !strings.Contains(suggestion, tt.wantSuggestion) {
					t.Errorf("ExplainExitCode(%d) suggestion = %q, want it to contain %q", tt.code, suggestion, tt.wantSuggestion)
				}
			}
		})
	}
}

func TestExplainExitCode_Signals(t *testing.T) {
	tests := []struct {
		code       int
		wantSignal string
	}{
		{-9, "SIGKILL"},
		{-15, "SIGTERM"},
	}
	for _, tt := range tests {
		t.Run(tt.wantSignal, func(t *testing.T) {
			meaning, _ := ExplainExitCode(tt.code)
			if !strings.Contains(meaning, tt.wantSignal) {
				t.Errorf("ExplainExitCode(%d) meaning = %q, want it to contain %q", tt.code, meaning, tt.wantSignal)
			}
			if !strings.Contains(meaning, "signal") && !strings.Contains(meaning, "Killed") {
				t.Errorf("ExplainExitCode(%d) meaning = %q, want it to mention signal/killed", tt.code, meaning)
			}
		})
	}
}

func TestExplainExitCode_UnknownCode(t *testing.T) {
	meaning, suggestion := ExplainExitCode(999)
	if meaning == "" {
		t.Error("ExplainExitCode(999) meaning should not be empty")
	}
	// Unknown code should still produce some meaningful output
	if !strings.Contains(meaning, "999") {
		t.Errorf("ExplainExitCode(999) meaning = %q, want it to contain the code number", meaning)
	}
	_ = suggestion // suggestion may or may not be empty for unknown codes
}

func TestExplainExitCode_UnknownNegative(t *testing.T) {
	meaning, _ := ExplainExitCode(-99)
	if meaning == "" {
		t.Error("ExplainExitCode(-99) meaning should not be empty")
	}
	if !strings.Contains(meaning, "signal") && !strings.Contains(meaning, "99") {
		t.Errorf("ExplainExitCode(-99) meaning = %q, want it to mention signal or code", meaning)
	}
}

func TestExplainExitCode_AllKnownCodes(t *testing.T) {
	// Ensure all documented codes produce non-empty meanings.
	knownCodes := []int{0, 1, 2, 5, 13, 78, 126, 127}
	for _, code := range knownCodes {
		meaning, _ := ExplainExitCode(code)
		if meaning == "" {
			t.Errorf("ExplainExitCode(%d) returned empty meaning", code)
		}
	}
}

func TestExplainExitCode_AllKnownSignals(t *testing.T) {
	knownSignals := []int{-1, -2, -9, -15}
	for _, code := range knownSignals {
		meaning, _ := ExplainExitCode(code)
		if meaning == "" {
			t.Errorf("ExplainExitCode(%d) returned empty meaning", code)
		}
	}
}
