package review

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/luuuc/council/internal/expert"
)

func testExpert() *expert.Expert {
	return &expert.Expert{
		ID:   "test-expert",
		Name: "Test Expert",
		Focus: "Testing",
		Body: "You are a testing expert.",
	}
}

func testSubmission() Submission {
	return Submission{Content: "func Add(a, b int) int { return a + b }"}
}

func TestAPIBackendAnthropic(t *testing.T) {
	verdictJSON := `{"expert":"test-expert","verdict":"pass","confidence":0.9,"notes":["Looks good"],"blocking":false}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request shape
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type")
		}
		if r.Header.Get("x-api-key") == "" {
			t.Errorf("expected x-api-key header")
		}
		if got := r.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Errorf("expected anthropic-version 2023-06-01, got %q", got)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["model"] != "claude-sonnet-4-6" {
			t.Errorf("expected model claude-sonnet-4-6, got %v", body["model"])
		}

		resp := map[string]any{
			"content": []map[string]string{
				{"type": "text", "text": verdictJSON},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("ANTHROPIC_API_KEY", "test-key-123")

	backend, err := newAPIBackendWithClient("anthropic", "claude-sonnet-4-6", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	backend.SetBaseURL(server.URL)

	verdict, err := backend.Review(context.Background(), testExpert(), testSubmission())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if verdict.Verdict != VerdictPass {
		t.Errorf("expected pass, got %s", verdict.Verdict)
	}
	if verdict.Confidence != 0.9 {
		t.Errorf("expected confidence 0.9, got %f", verdict.Confidence)
	}
	if len(verdict.Notes) != 1 || verdict.Notes[0] != "Looks good" {
		t.Errorf("unexpected notes: %v", verdict.Notes)
	}
}

func TestAPIBackendOpenAI(t *testing.T) {
	verdictJSON := `{"expert":"test-expert","verdict":"comment","confidence":0.7,"notes":["Add tests"],"blocking":false}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-openai-key" {
			t.Errorf("expected Bearer auth, got %q", got)
		}

		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": verdictJSON}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("OPENAI_API_KEY", "test-openai-key")

	backend, err := newAPIBackendWithClient("openai", "gpt-4o", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	backend.SetBaseURL(server.URL)

	verdict, err := backend.Review(context.Background(), testExpert(), testSubmission())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if verdict.Verdict != VerdictComment {
		t.Errorf("expected comment, got %s", verdict.Verdict)
	}
}

func TestAPIBackendOllama(t *testing.T) {
	verdictJSON := `{"expert":"test-expert","verdict":"block","confidence":0.85,"notes":["Missing validation"],"blocking":true}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ollama has no auth header
		if r.Header.Get("Authorization") != "" {
			t.Errorf("ollama should have no auth header")
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["stream"] != false {
			t.Errorf("expected stream: false for ollama")
		}

		resp := map[string]any{
			"message": map[string]string{"content": verdictJSON},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	backend, err := newAPIBackendWithClient("ollama", "llama3", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	backend.SetBaseURL(server.URL)

	verdict, err := backend.Review(context.Background(), testExpert(), testSubmission())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if verdict.Verdict != VerdictBlock {
		t.Errorf("expected block, got %s", verdict.Verdict)
	}
	if !verdict.Blocking {
		t.Error("expected blocking=true")
	}
}

func TestAPIBackendHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w,`{"error":{"message":"rate limit exceeded"}}`)
	}))
	defer server.Close()

	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	backend, err := newAPIBackendWithClient("anthropic", "claude-sonnet-4-6", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	backend.SetBaseURL(server.URL)

	_, err = backend.Review(context.Background(), testExpert(), testSubmission())
	if err == nil {
		t.Fatal("expected error for 429 response")
	}

	if got := err.Error(); !strings.Contains(got, "429") {
		t.Errorf("error should mention status code 429, got: %s", got)
	}
}

func TestAPIBackendServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w,`{"error":{"message":"internal server error"}}`)
	}))
	defer server.Close()

	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	backend, err := newAPIBackendWithClient("anthropic", "claude-sonnet-4-6", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	backend.SetBaseURL(server.URL)

	_, err = backend.Review(context.Background(), testExpert(), testSubmission())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestAPIBackendContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second) // simulate slow response
	}))
	defer server.Close()

	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	backend, err := newAPIBackendWithClient("anthropic", "claude-sonnet-4-6", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	backend.SetBaseURL(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = backend.Review(ctx, testExpert(), testSubmission())
	if err == nil {
		t.Fatal("expected error from context cancellation")
	}
}

func TestAPIBackendEmptyAnthropicResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"content": []any{}})
	}))
	defer server.Close()

	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	backend, err := newAPIBackendWithClient("anthropic", "claude-sonnet-4-6", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	backend.SetBaseURL(server.URL)

	_, err = backend.Review(context.Background(), testExpert(), testSubmission())
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestAPIBackendEmptyOpenAIResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{}})
	}))
	defer server.Close()

	t.Setenv("OPENAI_API_KEY", "test-key")

	backend, err := newAPIBackendWithClient("openai", "gpt-4o", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	backend.SetBaseURL(server.URL)

	_, err = backend.Review(context.Background(), testExpert(), testSubmission())
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestAPIBackendUnknownProvider(t *testing.T) {
	_, err := NewAPIBackend("gemini", "gemini-pro")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestAPIBackendMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w,"not json at all")
	}))
	defer server.Close()

	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	backend, err := newAPIBackendWithClient("anthropic", "claude-sonnet-4-6", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	backend.SetBaseURL(server.URL)

	_, err = backend.Review(context.Background(), testExpert(), testSubmission())
	if err == nil {
		t.Fatal("expected error for malformed JSON response")
	}
}

// TestAPIBackendRunnerIntegration verifies APIBackend works with the Runner.
func TestAPIBackendRunnerIntegration(t *testing.T) {
	verdictJSON := `{"expert":"test-expert","verdict":"pass","confidence":0.95,"notes":["Clean code"],"blocking":false}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"content": []map[string]string{
				{"type": "text", "text": verdictJSON},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	backend, err := newAPIBackendWithClient("anthropic", "claude-sonnet-4-6", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	backend.SetBaseURL(server.URL)

	runner := &Runner{
		Backend: backend,
		Options: ReviewOptions{Concurrency: 2, Timeout: 10},
	}

	inputs := []ExpertInput{
		{Expert: testExpert(), Blocking: false},
	}

	result := runner.Run(context.Background(), inputs, testSubmission())

	if len(result.Perspectives) != 1 {
		t.Fatalf("expected 1 perspective, got %d", len(result.Perspectives))
	}
	if result.Verdict != VerdictPass {
		t.Errorf("expected pass, got %s", result.Verdict)
	}
	if len(result.Errors) != 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}
