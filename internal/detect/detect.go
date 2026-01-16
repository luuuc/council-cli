package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Detection holds the detected project stack
type Detection struct {
	Languages  []Language  `json:"languages" yaml:"languages"`
	Frameworks []Framework `json:"frameworks" yaml:"frameworks"`
	Testing    []string    `json:"testing" yaml:"testing"`
	Patterns   []string    `json:"patterns" yaml:"patterns"`
}

// Language represents a detected language with percentage
type Language struct {
	Name       string  `json:"name" yaml:"name"`
	Percentage float64 `json:"percentage" yaml:"percentage"`
}

// Framework represents a detected framework with optional version
type Framework struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

// JSON returns the detection as JSON
func (d *Detection) JSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

// Summary returns a human-readable summary
func (d *Detection) Summary() string {
	var parts []string

	// Languages
	for _, lang := range d.Languages {
		parts = append(parts, lang.Name)
	}

	// Frameworks
	for _, fw := range d.Frameworks {
		if fw.Version != "" {
			parts = append(parts, fw.Name+" "+fw.Version)
		} else {
			parts = append(parts, fw.Name)
		}
	}

	// Testing
	parts = append(parts, d.Testing...)

	return strings.Join(parts, ", ")
}

// Scan analyzes the current directory and returns detection results
func Scan(dir string) (*Detection, error) {
	d := &Detection{
		Languages:  []Language{},
		Frameworks: []Framework{},
		Testing:    []string{},
		Patterns:   []string{},
	}

	// Count files by extension for language detection
	extCounts := make(map[string]int)
	totalFiles := 0

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden directories and common non-source directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") ||
				name == "node_modules" ||
				name == "vendor" ||
				name == "tmp" ||
				name == "log" ||
				name == "coverage" ||
				name == "dist" ||
				name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != "" && isSourceFile(ext) {
			extCounts[ext]++
			totalFiles++
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert extensions to languages
	langCounts := make(map[string]int)
	for ext, count := range extCounts {
		if lang := extToLanguage(ext); lang != "" {
			langCounts[lang] += count
		}
	}

	// Calculate percentages
	for lang, count := range langCounts {
		pct := 0.0
		if totalFiles > 0 {
			pct = float64(count) / float64(totalFiles) * 100
		}
		if pct >= 1.0 { // Only include languages >= 1%
			d.Languages = append(d.Languages, Language{
				Name:       lang,
				Percentage: float64(int(pct*10)) / 10, // Round to 1 decimal
			})
		}
	}

	// Sort languages by percentage descending
	sort.Slice(d.Languages, func(i, j int) bool {
		return d.Languages[i].Percentage > d.Languages[j].Percentage
	})

	// Detect frameworks
	d.detectFrameworks(dir)

	// Detect testing tools
	d.detectTesting(dir)

	// Detect patterns
	d.detectPatterns(dir)

	return d, nil
}

func isSourceFile(ext string) bool {
	sourceExts := map[string]bool{
		".go": true, ".rs": true, ".rb": true, ".py": true, ".js": true,
		".ts": true, ".tsx": true, ".jsx": true, ".vue": true, ".svelte": true,
		".java": true, ".kt": true, ".scala": true, ".swift": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".cs": true, ".fs": true, ".php": true, ".ex": true, ".exs": true,
		".erl": true, ".hrl": true, ".clj": true, ".cljs": true,
		".html": true, ".erb": true, ".haml": true, ".slim": true,
		".css": true, ".scss": true, ".sass": true, ".less": true,
	}
	return sourceExts[ext]
}

func extToLanguage(ext string) string {
	mapping := map[string]string{
		".go":     "Go",
		".rs":     "Rust",
		".rb":     "Ruby",
		".py":     "Python",
		".js":     "JavaScript",
		".ts":     "TypeScript",
		".tsx":    "TypeScript",
		".jsx":    "JavaScript",
		".vue":    "Vue",
		".svelte": "Svelte",
		".java":   "Java",
		".kt":     "Kotlin",
		".scala":  "Scala",
		".swift":  "Swift",
		".c":      "C",
		".cpp":    "C++",
		".cs":     "C#",
		".php":    "PHP",
		".ex":     "Elixir",
		".exs":    "Elixir",
		".erl":    "Erlang",
		".clj":    "Clojure",
		".erb":    "Ruby",
		".html":   "HTML",
		".css":    "CSS",
		".scss":   "CSS",
	}
	return mapping[ext]
}

func (d *Detection) detectFrameworks(dir string) {
	// Ruby/Rails
	if fileExists(dir, "Gemfile") {
		content := readFile(dir, "Gemfile")
		if strings.Contains(content, "rails") {
			version := extractGemVersion(content, "rails")
			d.Frameworks = append(d.Frameworks, Framework{Name: "Rails", Version: version})
		}
		if strings.Contains(content, "sinatra") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Sinatra"})
		}
	}

	// Node.js
	if fileExists(dir, "package.json") {
		content := readFile(dir, "package.json")
		var pkg map[string]interface{}
		if json.Unmarshal([]byte(content), &pkg) == nil {
			deps := mergeDeps(pkg)

			if _, ok := deps["next"]; ok {
				d.Frameworks = append(d.Frameworks, Framework{Name: "Next.js"})
			} else if _, ok := deps["react"]; ok {
				d.Frameworks = append(d.Frameworks, Framework{Name: "React"})
			}

			if _, ok := deps["vue"]; ok {
				d.Frameworks = append(d.Frameworks, Framework{Name: "Vue"})
			}

			if _, ok := deps["express"]; ok {
				d.Frameworks = append(d.Frameworks, Framework{Name: "Express"})
			}

			if _, ok := deps["@hotwired/turbo"]; ok {
				d.Patterns = append(d.Patterns, "Hotwire")
			}
			if _, ok := deps["@hotwired/stimulus"]; ok {
				if !contains(d.Patterns, "Hotwire") {
					d.Patterns = append(d.Patterns, "Stimulus")
				}
			}
		}
	}

	// Python
	if fileExists(dir, "requirements.txt") || fileExists(dir, "pyproject.toml") {
		var content string
		if fileExists(dir, "requirements.txt") {
			content = readFile(dir, "requirements.txt")
		} else {
			content = readFile(dir, "pyproject.toml")
		}

		if strings.Contains(content, "django") || strings.Contains(content, "Django") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Django"})
		}
		if strings.Contains(content, "flask") || strings.Contains(content, "Flask") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Flask"})
		}
		if strings.Contains(content, "fastapi") || strings.Contains(content, "FastAPI") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "FastAPI"})
		}
	}

	// Go
	if fileExists(dir, "go.mod") {
		content := readFile(dir, "go.mod")
		if strings.Contains(content, "gin-gonic/gin") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Gin"})
		}
		if strings.Contains(content, "labstack/echo") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Echo"})
		}
		if strings.Contains(content, "gofiber/fiber") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Fiber"})
		}
	}

	// Rust
	if fileExists(dir, "Cargo.toml") {
		content := readFile(dir, "Cargo.toml")
		if strings.Contains(content, "actix-web") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Actix Web"})
		}
		if strings.Contains(content, "rocket") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Rocket"})
		}
		if strings.Contains(content, "axum") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Axum"})
		}
	}

	// Swift
	if fileExists(dir, "Package.swift") {
		content := readFile(dir, "Package.swift")
		if strings.Contains(content, "Vapor") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Vapor"})
		}
	}

	// Kotlin/Android
	if fileExists(dir, "build.gradle.kts") || fileExists(dir, "build.gradle") {
		content := readFile(dir, "build.gradle.kts")
		if content == "" {
			content = readFile(dir, "build.gradle")
		}
		if strings.Contains(content, "android") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Android"})
		}
		if strings.Contains(content, "ktor") {
			d.Frameworks = append(d.Frameworks, Framework{Name: "Ktor"})
		}
	}
}

func (d *Detection) detectTesting(dir string) {
	// Ruby
	if fileExists(dir, "Gemfile") {
		content := readFile(dir, "Gemfile")
		if strings.Contains(content, "minitest") {
			d.Testing = append(d.Testing, "Minitest")
		}
		if strings.Contains(content, "rspec") {
			d.Testing = append(d.Testing, "RSpec")
		}
	}

	// Node.js
	if fileExists(dir, "package.json") {
		content := readFile(dir, "package.json")
		var pkg map[string]interface{}
		if json.Unmarshal([]byte(content), &pkg) == nil {
			deps := mergeDeps(pkg)

			if _, ok := deps["jest"]; ok {
				d.Testing = append(d.Testing, "Jest")
			}
			if _, ok := deps["mocha"]; ok {
				d.Testing = append(d.Testing, "Mocha")
			}
			if _, ok := deps["vitest"]; ok {
				d.Testing = append(d.Testing, "Vitest")
			}
			if _, ok := deps["playwright"]; ok {
				d.Testing = append(d.Testing, "Playwright")
			}
			if _, ok := deps["cypress"]; ok {
				d.Testing = append(d.Testing, "Cypress")
			}
		}
	}

	// Python
	if fileExists(dir, "requirements.txt") || fileExists(dir, "pyproject.toml") {
		var content string
		if fileExists(dir, "requirements.txt") {
			content = readFile(dir, "requirements.txt")
		} else {
			content = readFile(dir, "pyproject.toml")
		}

		if strings.Contains(content, "pytest") {
			d.Testing = append(d.Testing, "pytest")
		}
		if strings.Contains(content, "unittest") {
			d.Testing = append(d.Testing, "unittest")
		}
	}

	// Go - check for test files
	hasTests := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			hasTests = true
			return filepath.SkipAll
		}
		return nil
	})
	if hasTests && hasLanguage(d.Languages, "Go") {
		d.Testing = append(d.Testing, "Go testing")
	}

	// Rust
	if fileExists(dir, "Cargo.toml") {
		// Rust has built-in testing
		if hasLanguage(d.Languages, "Rust") {
			d.Testing = append(d.Testing, "Rust testing")
		}
	}
}

func (d *Detection) detectPatterns(dir string) {
	// Rails patterns
	if dirExists(dir, "app/controllers") {
		d.Patterns = append(d.Patterns, "MVC")
	}
	if dirExists(dir, "app/models/concerns") || dirExists(dir, "app/controllers/concerns") {
		d.Patterns = append(d.Patterns, "Concerns")
	}
	if dirExists(dir, "app/jobs") {
		d.Patterns = append(d.Patterns, "Background Jobs")
	}

	// API patterns
	if dirExists(dir, "app/api") || dirExists(dir, "src/api") || dirExists(dir, "api") {
		d.Patterns = append(d.Patterns, "API")
	}

	// Monorepo
	if dirExists(dir, "packages") || fileExists(dir, "pnpm-workspace.yaml") {
		d.Patterns = append(d.Patterns, "Monorepo")
	}

	// Microservices
	if dirExists(dir, "services") && !dirExists(dir, "app/services") {
		d.Patterns = append(d.Patterns, "Microservices")
	}

	// Check for Tailwind
	if fileExists(dir, "tailwind.config.js") || fileExists(dir, "tailwind.config.ts") {
		d.Patterns = append(d.Patterns, "Tailwind CSS")
	}

	// Check for Docker
	if fileExists(dir, "Dockerfile") || fileExists(dir, "docker-compose.yml") || fileExists(dir, "docker-compose.yaml") {
		d.Patterns = append(d.Patterns, "Docker")
	}

	// Check for CI/CD
	if dirExists(dir, ".github/workflows") {
		d.Patterns = append(d.Patterns, "GitHub Actions")
	}
}

// Helper functions

func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func dirExists(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && info.IsDir()
}

func readFile(dir, name string) string {
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return ""
	}
	return string(data)
}

func extractGemVersion(content, gem string) string {
	// Simple version extraction from Gemfile
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, `"`+gem+`"`) || strings.Contains(line, `'`+gem+`'`) {
			// Look for version like "~> 8.0" or "7.1.0"
			if idx := strings.Index(line, ","); idx > 0 {
				rest := line[idx+1:]
				rest = strings.TrimSpace(rest)
				rest = strings.Trim(rest, `"'`)
				if strings.HasPrefix(rest, "~>") || strings.HasPrefix(rest, ">=") {
					rest = strings.TrimPrefix(rest, "~>")
					rest = strings.TrimPrefix(rest, ">=")
					rest = strings.TrimSpace(rest)
					rest = strings.Trim(rest, `"'`)
					return rest
				}
			}
		}
	}
	return ""
}

func mergeDeps(pkg map[string]interface{}) map[string]interface{} {
	deps := make(map[string]interface{})
	if d, ok := pkg["dependencies"].(map[string]interface{}); ok {
		for k, v := range d {
			deps[k] = v
		}
	}
	if d, ok := pkg["devDependencies"].(map[string]interface{}); ok {
		for k, v := range d {
			deps[k] = v
		}
	}
	return deps
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func hasLanguage(langs []Language, name string) bool {
	for _, l := range langs {
		if l.Name == name {
			return true
		}
	}
	return false
}
