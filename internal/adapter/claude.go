package adapter

import (
	"fmt"
	"os"
	"strings"

	"github.com/luuuc/council-cli/internal/expert"
)

func init() {
	Register(&Claude{})
}

// Claude is the adapter for Claude Code.
type Claude struct{}

func (c *Claude) Name() string {
	return "claude"
}

func (c *Claude) DisplayName() string {
	return "Claude Code"
}

func (c *Claude) Detect() bool {
	return DirExists(".claude")
}

func (c *Claude) Paths() Paths {
	return Paths{
		Agents:     ".claude/agents",
		Commands:   ".claude/commands",
		Deprecated: []string{},
	}
}

func (c *Claude) Templates() Templates {
	return Templates{
		Install:  claudeInstallTemplate,
		Commands: claudeCommandTemplates,
	}
}

// FormatAgent creates Claude Code agent file content.
// For Claude Code, we use the original expert file content (preserves source format).
func (c *Claude) FormatAgent(e *expert.Expert) string {
	// Read the original expert file and return its content
	data, err := os.ReadFile(e.Path())
	if err != nil {
		// Fallback to regenerating
		return fmt.Sprintf("---\nid: %s\nname: %s\nfocus: %s\n---\n\n%s", e.ID, e.Name, e.Focus, e.Body)
	}
	return string(data)
}

// FormatCommand creates Claude Code command file content.
// Claude Code commands are plain markdown (no frontmatter needed).
func (c *Claude) FormatCommand(name, description, body string) string {
	return body
}

// Claude Code templates

var claudeInstallTemplate = `# Install Council

Your AI tool will read the appropriate instructions for setting up the council.

## Quick Start

1. Check if council is already initialized:
` + "```bash" + `
ls -la .council/
` + "```" + `

2. If not initialized, run:
` + "```bash" + `
council init
` + "```" + `

3. Analyze your project and suggest experts:
` + "```bash" + `
council detect --json
` + "```" + `

4. Add experts to your council - use one of:
   - ` + "`/council-add`" + ` - interactive expert search
   - ` + "`council add \"Expert Name\"`" + ` - add specific curated expert
   - ` + "`council setup --apply`" + ` - AI-assisted setup

5. Sync to your AI tool:
` + "```bash" + `
council sync
` + "```"

var claudeCommandTemplates = map[string]string{
	"council-add": `# Add Expert to Council

Add a new expert to the council: $ARGUMENTS

## Step 1: Classify Input

Determine what type of input $ARGUMENTS is:

| Type | Pattern | Examples |
|------|---------|----------|
| **Name** | Quoted string, or 2-3 capitalized words forming a person's name | ` + "`\"Kent Beck\"`" + `, ` + "`Sandi Metz`" + ` |
| **Description** | Contains "a ", "someone", "expert in", "help with" | ` + "`a testing expert`" + `, ` + "`someone for API design`" + ` |
| **Keyword** | Single word describing a domain | ` + "`testing`" + `, ` + "`APIs`" + `, ` + "`security`" + ` |

**If ambiguous, treat as description and search.**

## Step 2: Search Curated Personas

Run this command to get all available curated personas:
` + "```bash" + `
council personas --json
` + "```" + `

Search the output for matches against:
- ` + "`name`" + ` and ` + "`id`" + ` - for name inputs (exact or close match)
- ` + "`focus`" + ` - primary match for descriptions/keywords
- ` + "`triggers`" + ` - semantic keywords that indicate relevance
- ` + "`philosophy`" + ` - deeper concept matching

Rank matches by relevance. A persona matching multiple fields ranks higher.

Also check who is already installed:
` + "```bash" + `
council list --json
` + "```" + `

Avoid suggesting experts already in the council.

## Step 3: Build 4 Suggestions (Rule of 4)

Always present exactly 4 options. This provides enough choice without overwhelming.

**If Name input found in curated personas:**
Skip to the Name Found flow (Step 4a).

**If Name input NOT found in curated personas:**
Skip to the Name Not Found flow (Step 4b).

**If Description or Keyword input:**

Count curated matches and build 4 options:

**0 curated matches:**
- Slots 1-3: AI-suggested well-known experts in the domain
- Slot 4: "Custom" - create persona from description

**1-3 curated matches:**
- Slots 1-N: Curated matches (most relevant first)
- Remaining slots: AI-suggested well-known experts to complement
- Slot 4: "Custom" (always last)

**4+ curated matches:**
- Slots 1-3: Top 3 curated matches (most relevant)
- Slot 4: "Show more curated" or "Custom"

## Step 4: Present Options

### Step 4a: Name Found in Curated

If the user provided a name and it matches a curated persona, tell the user what you found, then use **AskUserQuestion** with these options:

| Label | Description |
|-------|-------------|
| "Add {Name}" | Install this expert from the curated library |
| "Show alternatives" | See related experts before deciding |
| "Cancel" | Don't add anyone |

- If **Add {Name}**: Run ` + "`council add \"{name}\"`" + ` to install the curated persona.
- If **Show alternatives**: Present 4 options including the match and related experts.

### Step 4b: Name Not Found

If the user provided a name but it's not in the curated library, tell the user the name wasn't found, then use **AskUserQuestion** with these options:

| Label | Description |
|-------|-------------|
| "Create {Name}" | I'll research and build a custom profile |
| "Show similar" | Browse related curated personas |
| "Cancel" | Don't add anyone |

- If **Create {Name}**: Proceed to Step 5 to generate a custom profile.
- If **Show similar**: Search curated personas by inferred domain and present 4 options.

### Step 4c: Description or Keyword

For description or keyword inputs, build exactly 4 options and use **AskUserQuestion** to present them:

| Label | Description |
|-------|-------------|
| "{Name}" | {focus} - from curated library |
| "{Name}" | {focus} - from curated library |
| "{Name}" | {brief expertise} - will create custom profile |
| "Custom" | Create a persona matching your description |

Labels should indicate source:
- **(curated)** options use ` + "`council add \"{name}\"`" + ` - has pre-written philosophy/principles
- **(suggested)** options need custom profile generation

When the user selects:
- **Curated option**: Run ` + "`council add \"{name}\"`" + ` to install
- **Suggested option**: Proceed to Step 5 to generate a custom profile
- **Custom**: Ask for any additional details, then proceed to Step 5

## Step 5: Generate Custom Profile

Once an expert is confirmed for custom creation, generate a rich profile.

Research or use your knowledge of this person to generate:

1. **Philosophy** (2-4 sentences): What they believe about software/design. Write in first person as if they're speaking. Capture their distinctive worldview.

2. **Principles** (4-6 items): Concrete, actionable guidelines they're known for. These should be memorable and specific to their thinking.

3. **Red Flags** (3-5 items): Patterns they would call out during code review. Things that violate their principles.

### Output Format

Create the expert file at ` + "`.council/experts/{id}.md`" + ` with this structure:

` + "```markdown" + `
---
id: {kebab-case-id}
name: {Full Name}
focus: {focus area}
philosophy: |
  {philosophy text - first person, 2-4 sentences}
principles:
  - {principle 1}
  - {principle 2}
  - {principle 3}
  - {principle 4}
red_flags:
  - {red flag 1}
  - {red flag 2}
  - {red flag 3}
---

# {Name} - {focus}

You are channeling {Name}, known for expertise in {focus}.

## Philosophy

{philosophy text}

## Principles

- {principle 1}
- {principle 2}
- ...

## Red Flags

Watch for these patterns:
- {red flag 1}
- {red flag 2}
- ...

## Review Style

{2-3 sentences describing how they approach code review}
` + "```" + `

## After Creating

1. Write the file using your file writing capability
2. Run ` + "`council sync`" + ` to update AI tool configurations
3. Confirm creation with: "Added {Name} ({id}) to the council"
4. Show the file path

**Note:** For curated personas, use ` + "`council add \"{name}\"`" + ` instead of writing files directly.
`,

	"council-detect": `# Detect Stack and Suggest Experts

Analyze this codebase and suggest council experts.

## Step 1: Run Detection

First, run the detection command to get structured stack information:

` + "```bash" + `
council detect --json
` + "```" + `

This gives you the detected languages, frameworks, testing tools, and patterns.

## Step 2: Analyze Results

Review the detection output and supplement with your codebase knowledge:

1. **Languages**: What's the primary language by percentage?
2. **Frameworks**: What frameworks are detected?
3. **Testing**: What testing tools/approaches are in use?
4. **Patterns**: What architectural patterns are detected?
5. **Domain**: What problem domain does this project address? (Use your codebase context)

## Step 3: Suggest Experts

Based on the detection, suggest **3-5 experts** (maximum 7) who would be valuable council members. Consider:

- **Framework experts** (1-2): DHH for Rails, Chris McCord for Phoenix, etc.
- **Language experts** (1-2): Rob Pike for Go, Matz for Ruby, etc.
- **Practice experts** (1-2): Kent Beck if tests detected, Sandi Metz for OO design

**Selection criteria:**
- Each expert fills a unique role (no redundancy)
- Prioritize direct stack matches over general wisdom
- Quality over quantity

For each suggested expert, provide:
- **Name**: The expert's name
- **Focus**: Their specific focus area relevant to THIS project
- **Why**: One sentence on why they'd be valuable for this codebase

## Output Format

Present your findings:

` + "```" + `
## Detected Stack

**Primary Language**: {language} ({percentage}%)
**Frameworks**: {list}
**Testing**: {tools/approaches}
**Patterns**: {observed patterns}

## Suggested Council (3-5 experts)

1. **{Name}** - {Focus}
   Why: {reason specific to this codebase}

2. **{Name}** - {Focus}
   Why: {reason specific to this codebase}

3. **{Name}** - {Focus}
   Why: {reason specific to this codebase}
` + "```" + `

## After Analysis

Ask the user which experts they want to add. For curated experts (from ` + "`council personas --json`" + `), use ` + "`council add \"{Name}\"`" + `. For custom experts or discovery, use ` + "`/council-add`" + ` with a description:
` + "```" + `
/council-add a testing expert
/council-add someone for API design
` + "```" + `
The skill searches curated personas first and always presents 4 options.

If the user says "all" or "add them", add all suggested experts.
`,

	"council-remove": `# Remove Expert from Council

Remove an expert from the council: $ARGUMENTS

## Instructions

You are removing an expert from the council.

## Step 1: Identify the Expert

Parse the arguments to get the expert name or ID. If not provided, list current experts:

` + "```bash" + `
council list
` + "```" + `

Then ask the user which expert to remove.

## Step 2: Remove the Expert

Run the council remove command with the expert ID:

` + "```bash" + `
council remove {expert-id}
` + "```" + `

The command will ask for confirmation before removing.

## Step 3: Sync Changes

After removal, sync the changes to AI tool configurations:

` + "```bash" + `
council sync
` + "```" + `

## After Removing

Confirm with: "Removed {Name} from the council"
`,
}

// CouncilCommandTemplate is exported for use by sync when generating the dynamic /council command
func CouncilCommandTemplate() string {
	return `# Code Review Council

Convene the council to review: $ARGUMENTS

## Council Members

{{range .}}
### {{.Name}}
**Focus**: {{.Focus}}
{{end}}

## Instructions

Review the code from each expert's perspective. For each expert:
1. State the expert's name
2. Provide their assessment focused on their domain
3. Note any concerns or suggestions

At the end, synthesize the key points and provide actionable recommendations.
`
}

// agentFilename returns the appropriate filename for an expert based on source
// This is exported for use by sync package
func AgentFilename(e *expert.Expert) string {
	switch {
	case e.Source == "custom":
		return "custom-" + e.ID + ".md"
	case strings.HasPrefix(e.Source, "installed:"):
		return "installed-" + e.ID + ".md"
	default:
		return e.ID + ".md"
	}
}
