package detect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScan_DetectsRailsProjectFromGemfile(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "app.rb"), []byte("class App\nend"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "test.rb"), []byte("class Test\nend"), 0644)

	gemfile := `source 'https://rubygems.org'
gem "rails", "~> 8.0"
gem "minitest"
`
	_ = os.WriteFile(filepath.Join(dir, "Gemfile"), []byte(gemfile), 0644)

	d, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(d.Languages) == 0 {
		t.Error("Expected at least one language")
	}

	foundRuby := false
	for _, lang := range d.Languages {
		if lang.Name == "Ruby" {
			foundRuby = true
			break
		}
	}
	if !foundRuby {
		t.Error("Expected Ruby to be detected")
	}

	foundRails := false
	for _, fw := range d.Frameworks {
		if fw.Name == "Rails" {
			foundRails = true
			if fw.Version != "8.0" {
				t.Errorf("Expected Rails version 8.0, got %s", fw.Version)
			}
			break
		}
	}
	if !foundRails {
		t.Error("Expected Rails to be detected")
	}

	foundMinitest := false
	for _, test := range d.Testing {
		if test == "Minitest" {
			foundMinitest = true
			break
		}
	}
	if !foundMinitest {
		t.Error("Expected Minitest to be detected")
	}
}

func TestScan_ReturnsEmptyForEmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	d, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(d.Languages) != 0 {
		t.Errorf("Expected no languages in empty dir, got %d", len(d.Languages))
	}
}

func TestSummary_CombinesLanguagesFrameworksTesting(t *testing.T) {
	d := &Detection{
		Languages: []Language{
			{Name: "Ruby", Percentage: 90},
			{Name: "JavaScript", Percentage: 10},
		},
		Frameworks: []Framework{
			{Name: "Rails", Version: "8.0"},
		},
		Testing: []string{"Minitest"},
	}

	summary := d.Summary()
	if summary == "" {
		t.Error("Expected non-empty summary")
	}

	// Check that summary contains key elements
	if !strings.Contains(summary, "Ruby") {
		t.Error("Expected summary to contain Ruby")
	}
	if !strings.Contains(summary, "Rails 8.0") {
		t.Error("Expected summary to contain Rails 8.0")
	}
}
