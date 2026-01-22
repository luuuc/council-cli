package adapter

import (
	"fmt"
	"strings"

	"github.com/luuuc/council-cli/internal/expert"
)

func init() {
	Register(&OpenCode{})
}

// OpenCode is the adapter for OpenCode.
type OpenCode struct{}

func (o *OpenCode) Name() string {
	return "opencode"
}

func (o *OpenCode) DisplayName() string {
	return "OpenCode"
}

func (o *OpenCode) Detect() bool {
	return DirExists(".opencode") || FileExists("opencode.json")
}

func (o *OpenCode) Paths() Paths {
	return Paths{
		Agents:     ".opencode/agents",
		Commands:   ".opencode/commands",
		Deprecated: []string{".opencode/agent"}, // Old singular path
	}
}

func (o *OpenCode) Templates() Templates {
	return Templates{
		Install:  opencodeInstallTemplate,
		Commands: opencodeCommandTemplates,
	}
}

// FormatAgent creates OpenCode agent file content.
// OpenCode uses a different frontmatter format with description and mode.
func (o *OpenCode) FormatAgent(e *expert.Expert) string {
	var parts []string

	// OpenCode uses different frontmatter format
	parts = append(parts, "---")
	parts = append(parts, fmt.Sprintf("description: %s", e.Focus))
	parts = append(parts, "mode: subagent")
	parts = append(parts, "---")
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("# %s", e.Name))
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("You are channeling %s, known for expertise in %s.", e.Name, e.Focus))
	parts = append(parts, "")

	if e.Philosophy != "" {
		parts = append(parts, "## Philosophy")
		parts = append(parts, "")
		parts = append(parts, strings.TrimSpace(e.Philosophy))
		parts = append(parts, "")
	}

	if len(e.Principles) > 0 {
		parts = append(parts, "## Principles")
		parts = append(parts, "")
		for _, p := range e.Principles {
			parts = append(parts, fmt.Sprintf("- %s", p))
		}
		parts = append(parts, "")
	}

	if len(e.RedFlags) > 0 {
		parts = append(parts, "## Red Flags")
		parts = append(parts, "")
		parts = append(parts, "Watch for these patterns:")
		for _, r := range e.RedFlags {
			parts = append(parts, fmt.Sprintf("- %s", r))
		}
		parts = append(parts, "")
	}

	parts = append(parts, "## Review Style")
	parts = append(parts, "")
	parts = append(parts, "When reviewing code, focus on your area of expertise. Be direct and specific.")
	parts = append(parts, "Explain your reasoning. Suggest concrete improvements.")

	return strings.Join(parts, "\n")
}

// FormatCommand creates OpenCode command file content.
// OpenCode commands have frontmatter with description and mode.
func (o *OpenCode) FormatCommand(name, description, body string) string {
	var parts []string
	parts = append(parts, "---")
	parts = append(parts, fmt.Sprintf("description: %s", description))
	parts = append(parts, "mode: subagent")
	parts = append(parts, "---")
	parts = append(parts, "")
	parts = append(parts, body)
	return strings.Join(parts, "\n")
}

// OpenCode templates - use text-based choices instead of AskUserQuestion

var opencodeInstallTemplate = `# Install Council

Set up the council for your project.

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

var opencodeCommandTemplates = map[string]string{
	"council-add": `# Add Expert to Council

Add a new expert to the council: $ARGUMENTS

## Step 1: Classify Input

Determine what type of input $ARGUMENTS is:

- **Name**: Quoted string, or 2-3 capitalized words forming a person's name (e.g., "Kent Beck", Sandi Metz)
- **Description**: Contains "a ", "someone", "expert in", "help with" (e.g., "a testing expert")
- **Keyword**: Single word describing a domain (e.g., "testing", "APIs")

**If ambiguous, treat as description and search.**

## Step 2: Search Curated Personas

Run this command to get all available curated personas:
` + "```bash" + `
council personas --json
` + "```" + `

Search the output for matches against name, id, focus, triggers, and philosophy.
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

If the user provided a name and it matches a curated persona, present these options:

` + "```" + `
Found "{Name}" in the curated library.

Options:
1. Add {Name} - Install this expert from the curated library
2. Show alternatives - See related experts before deciding
3. Cancel - Don't add anyone

Which option? (1/2/3):
` + "```" + `

Wait for the user's response, then:
- Option 1: Run ` + "`council add \"{name}\"`" + ` to install
- Option 2: Present 4 options including the match and related experts
- Option 3: Exit

### Step 4b: Name Not Found

If the user provided a name but it's not in the curated library:

` + "```" + `
"{Name}" is not in the curated library.

Options:
1. Create {Name} - I'll research and build a custom profile
2. Show similar - Browse related curated personas
3. Cancel - Don't add anyone

Which option? (1/2/3):
` + "```" + `

Wait for the user's response, then:
- Option 1: Proceed to Step 5 to generate a custom profile
- Option 2: Search curated personas by inferred domain and present 4 options
- Option 3: Exit

### Step 4c: Description or Keyword

For description or keyword inputs, present 4 options:

` + "```" + `
Based on your request, here are 4 options:

1. {Name} - {focus} (curated)
2. {Name} - {focus} (curated)
3. {Name} - {brief expertise} (will create custom profile)
4. Custom - Create a persona matching your description

Which option? (1/2/3/4):
` + "```" + `

Wait for the user's response, then:
- Curated options: Run ` + "`council add \"{name}\"`" + ` to install
- Suggested options: Proceed to Step 5 to generate a custom profile
- Custom: Ask for additional details, then proceed to Step 5

## Step 5: Generate Custom Profile

Once an expert is confirmed for custom creation, generate a rich profile.

Research or use your knowledge of this person to generate:

1. **Philosophy** (2-4 sentences): What they believe about software/design. Write in first person.
2. **Principles** (4-6 items): Concrete, actionable guidelines they're known for.
3. **Red Flags** (3-5 items): Patterns they would call out during code review.

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
5. **Domain**: What problem domain does this project address?

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

Ask the user which experts they want to add. For curated experts, use ` + "`council add \"{Name}\"`" + `.
For custom experts, use ` + "`/council-add`" + ` with a description.

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
