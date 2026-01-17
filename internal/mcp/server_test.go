package mcp

import (
	"context"
	"os"
	"testing"

	"github.com/luuuc/council-cli/internal/config"
	"github.com/luuuc/council-cli/internal/expert"
	"github.com/mark3labs/mcp-go/mcp"
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
		t.Fatal("NewServer() returned nil")
	}
	if s.mcp == nil {
		t.Error("NewServer().mcp is nil")
	}
}

func setupTestCouncil(t *testing.T) (cleanup func()) {
	t.Helper()

	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "council-mcp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)

	// Create the council directory structure
	_ = os.MkdirAll(config.Path(config.ExpertsDir), 0755)

	// Create config
	cfg := config.Default()
	_ = cfg.Save()

	return func() {
		_ = os.Chdir(origDir)
		os.RemoveAll(tmpDir)
	}
}

func TestHandleListExperts_NoCouncil(t *testing.T) {
	// Create temp dir without council
	tmpDir, err := os.MkdirTemp("", "council-mcp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	s := NewServer()
	ctx := context.Background()
	req := mcp.CallToolRequest{}

	result, err := s.handleListExperts(ctx, req)
	if err != nil {
		t.Errorf("handleListExperts() error = %v", err)
	}
	if result == nil {
		t.Fatal("handleListExperts() returned nil result")
	}
	if !result.IsError {
		t.Error("handleListExperts() should return error when council not initialized")
	}
}

func TestHandleListExperts_NoExperts(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	s := NewServer()
	ctx := context.Background()
	req := mcp.CallToolRequest{}

	result, err := s.handleListExperts(ctx, req)
	if err != nil {
		t.Errorf("handleListExperts() error = %v", err)
	}
	if result == nil {
		t.Fatal("handleListExperts() returned nil result")
	}
	if result.IsError {
		t.Error("handleListExperts() should not error when council exists but is empty")
	}
}

func TestHandleListExperts_WithExperts(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	// Create test expert
	testExpert := &expert.Expert{
		ID:    "test-expert",
		Name:  "Test Expert",
		Focus: "Testing MCP handlers",
	}
	_ = testExpert.Save()

	s := NewServer()
	ctx := context.Background()
	req := mcp.CallToolRequest{}

	result, err := s.handleListExperts(ctx, req)
	if err != nil {
		t.Errorf("handleListExperts() error = %v", err)
	}
	if result == nil {
		t.Fatal("handleListExperts() returned nil result")
	}
	if result.IsError {
		t.Error("handleListExperts() should not error with valid experts")
	}
}

func TestHandleGetExpert_NotFound(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	s := NewServer()
	ctx := context.Background()

	// Create a request with an ID parameter
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id": "nonexistent",
			},
		},
	}

	result, err := s.handleGetExpert(ctx, req)
	if err != nil {
		t.Errorf("handleGetExpert() error = %v", err)
	}
	if result == nil {
		t.Fatal("handleGetExpert() returned nil result")
	}
	if !result.IsError {
		t.Error("handleGetExpert() should return error for non-existent expert")
	}
}

func TestHandleGetExpert_Found(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	// Create test expert
	testExpert := &expert.Expert{
		ID:         "test-expert",
		Name:       "Test Expert",
		Focus:      "Testing",
		Philosophy: "Test everything.",
		Principles: []string{"Write tests first"},
		RedFlags:   []string{"No tests"},
	}
	_ = testExpert.Save()

	s := NewServer()
	ctx := context.Background()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"id": "test-expert",
			},
		},
	}

	result, err := s.handleGetExpert(ctx, req)
	if err != nil {
		t.Errorf("handleGetExpert() error = %v", err)
	}
	if result == nil {
		t.Fatal("handleGetExpert() returned nil result")
	}
	if result.IsError {
		t.Error("handleGetExpert() should not error for existing expert")
	}
}

func TestHandleConsultCouncil_NoExperts(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	s := NewServer()
	ctx := context.Background()
	req := mcp.CallToolRequest{}

	result, err := s.handleConsultCouncil(ctx, req)
	if err != nil {
		t.Errorf("handleConsultCouncil() error = %v", err)
	}
	if result == nil {
		t.Fatal("handleConsultCouncil() returned nil result")
	}
	// Empty council should return text, not error
	if result.IsError {
		t.Error("handleConsultCouncil() should not error for empty council")
	}
}

func TestHandleConsultCouncil_WithExperts(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	// Create test experts
	expert1 := &expert.Expert{
		ID:    "expert-1",
		Name:  "Expert One",
		Focus: "Area One",
	}
	_ = expert1.Save()

	expert2 := &expert.Expert{
		ID:    "expert-2",
		Name:  "Expert Two",
		Focus: "Area Two",
	}
	_ = expert2.Save()

	s := NewServer()
	ctx := context.Background()
	req := mcp.CallToolRequest{}

	result, err := s.handleConsultCouncil(ctx, req)
	if err != nil {
		t.Errorf("handleConsultCouncil() error = %v", err)
	}
	if result == nil {
		t.Fatal("handleConsultCouncil() returned nil result")
	}
	if result.IsError {
		t.Error("handleConsultCouncil() should not error with experts")
	}
}

func TestHandleExpertResource_InvalidURI(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	s := NewServer()
	ctx := context.Background()

	req := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "invalid-uri",
		},
	}

	_, err := s.handleExpertResource(ctx, req)
	if err == nil {
		t.Error("handleExpertResource() should error for invalid URI")
	}
}

func TestHandleExpertResource_NotFound(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	s := NewServer()
	ctx := context.Background()

	req := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "council://experts/nonexistent",
		},
	}

	_, err := s.handleExpertResource(ctx, req)
	if err == nil {
		t.Error("handleExpertResource() should error for non-existent expert")
	}
}

func TestHandleExpertResource_Found(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	// Create test expert
	testExpert := &expert.Expert{
		ID:    "test-expert",
		Name:  "Test Expert",
		Focus: "Testing",
		Body:  "# Test Expert\n\nBody content.",
	}
	_ = testExpert.Save()

	s := NewServer()
	ctx := context.Background()

	req := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "council://experts/test-expert",
		},
	}

	contents, err := s.handleExpertResource(ctx, req)
	if err != nil {
		t.Errorf("handleExpertResource() error = %v", err)
	}
	if len(contents) != 1 {
		t.Errorf("handleExpertResource() returned %d contents, want 1", len(contents))
	}
}

func TestHandleCouncilPrompt_EmptyContent(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	s := NewServer()
	ctx := context.Background()

	req := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name:      "council",
			Arguments: map[string]string{},
		},
	}

	result, err := s.handleCouncilPrompt(ctx, req)
	if err != nil {
		t.Errorf("handleCouncilPrompt() error = %v", err)
	}
	if result == nil {
		t.Fatal("handleCouncilPrompt() returned nil result")
	}
	if result.Description == "" {
		t.Error("handleCouncilPrompt() should have description")
	}
}

func TestHandleCouncilPrompt_WithContent(t *testing.T) {
	cleanup := setupTestCouncil(t)
	defer cleanup()

	// Create test expert
	testExpert := &expert.Expert{
		ID:    "test-expert",
		Name:  "Test Expert",
		Focus: "Testing",
	}
	_ = testExpert.Save()

	s := NewServer()
	ctx := context.Background()

	req := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "council",
			Arguments: map[string]string{
				"content": "Review this code please",
			},
		},
	}

	result, err := s.handleCouncilPrompt(ctx, req)
	if err != nil {
		t.Errorf("handleCouncilPrompt() error = %v", err)
	}
	if result == nil {
		t.Fatal("handleCouncilPrompt() returned nil result")
	}
	if len(result.Messages) == 0 {
		t.Error("handleCouncilPrompt() should return messages")
	}
}
