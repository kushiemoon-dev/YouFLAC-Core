package core

import "testing"

func TestDetectExplicit(t *testing.T) {
	tests := []struct {
		title string
		want  bool
	}{
		{"Some Song [Explicit]", true},
		{"Some Song (Explicit)", true},
		{"Some Song - Explicit Version", true},
		{"Some Song [explicit]", true},
		{"Explicitly Yours", false},
		{"Clean Version", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			if got := DetectExplicit(tc.title); got != tc.want {
				t.Errorf("DetectExplicit(%q) = %v, want %v", tc.title, got, tc.want)
			}
		})
	}
}
