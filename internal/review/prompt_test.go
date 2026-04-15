package review

import (
	"strings"
	"testing"

	"github.com/luuuc/council-cli/internal/expert"
)

func TestBuildPrompt(t *testing.T) {
	e := &expert.Expert{
		ID:   "kent-beck",
		Name: "Kent Beck",
		Body: "# Kent Beck - Test-driven development\n\nYou are channeling Kent Beck.",
	}

	sub := Submission{
		Content: "diff --git a/main.go\n+func Add(a, b int) int {\n+    return a + b\n+}",
		Context: "PR: Add math utilities",
	}

	prompt := BuildPrompt(e, sub)

	checks := []string{
		"You are Kent Beck",
		"Kent Beck - Test-driven development",
		"diff --git a/main.go",
		"PR: Add math utilities",
		`"verdict"`,
		`"kent-beck"`,
		"pass|comment|block|escalate",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("prompt missing %q\n\nPrompt:\n%s", check, prompt)
		}
	}
}

func TestBuildPromptNoContext(t *testing.T) {
	e := &expert.Expert{
		ID:   "kent-beck",
		Name: "Kent Beck",
		Body: "# Kent Beck\n\nExpert.",
	}

	sub := Submission{Content: "some code"}

	prompt := BuildPrompt(e, sub)

	if strings.Contains(prompt, "## Context") {
		t.Error("prompt should not contain Context section when context is empty")
	}
}
