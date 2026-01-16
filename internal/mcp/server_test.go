package mcp

import (
	"testing"
)

func TestExtractExpertID(t *testing.T) {
	tests := []struct {
		uri      string
		expected string
	}{
		{"council://experts/dhh", "dhh"},
		{"council://experts/kent-beck", "kent-beck"},
		{"council://experts/", ""},
		{"council://experts", ""},
		{"invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			result := extractExpertID(tt.uri)
			if result != tt.expected {
				t.Errorf("extractExpertID(%q) = %q, want %q", tt.uri, result, tt.expected)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	s := NewServer()
	if s == nil {
		t.Error("NewServer() returned nil")
	}
	if s.mcp == nil {
		t.Error("NewServer().mcp is nil")
	}
}
