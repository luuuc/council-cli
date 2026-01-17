# council-cli

AI-agnostic expert council setup for coding assistants.

## 30-Second Start

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/luuuc/council-cli/main/install.sh | sh

# Create your council (interactive, no AI needed)
council init && council setup -i

# Sync to your AI tool
council sync
```

Your AI now has an expert council. Try `/council` in Claude Code or Cursor.

## What is a Council?

Tired of generic AI code reviews? The council pattern establishes **expert personas** that represent excellence in specific domains:

```
Your Project           council setup              AI Config Files
┌────────────┐         ┌──────────┐              ┌────────────┐
│ Rails 8    │         │          │              │ .claude/   │
│ Hotwire    │────────▶│ suggests │─────────────▶│ .cursor/   │
│ Minitest   │         │ experts  │              │ .windsurf/ │
│ SQLite     │         │          │              │ AGENTS.md  │
└────────────┘         └──────────┘              └────────────┘
```

Instead of generic AI review, you get specific expertise in the domains that matter for your project.

> **Beyond code**: The council pattern works for any domain—marketing strategy, writing, business decisions, philosophy. If you want AI to think like your favorite experts, council-cli can help.

## Setup Modes

Three ways to create your council:

### Interactive (Recommended for first-timers)

Built-in suggestions based on your tech stack. Works offline, no external AI needed.

```bash
council setup -i
```

### AI-Assisted

Let an external AI CLI suggest experts tailored to your project.

```bash
council setup --apply        # Uses configured AI (claude, aichat, llm, etc.)
council setup --apply --yes  # Skip confirmation
```

### Manual

Generate a prompt to paste into any AI, then pipe the response back.

```bash
council setup                    # Output prompt to copy
council setup -o prompt.md       # Save prompt to file
council setup --apply < response.yaml  # Apply AI response
```

## Installation

```bash
# Direct download (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/luuuc/council-cli/main/install.sh | sh

# Go
go install github.com/luuuc/council-cli/cmd/council@latest

# From source
git clone https://github.com/luuuc/council-cli.git
cd council-cli
make install
```

## Integrations

### Claude Code / Cursor / Windsurf

After `council sync`, your experts are available as slash commands:

```bash
council sync           # Sync to all configured targets
council sync claude    # Sync to specific target
council sync --dry-run # Preview changes
```

| Target | Location | Description |
|--------|----------|-------------|
| `claude` | `.claude/agents/`, `.claude/commands/` | Claude Code |
| `cursor` | `.cursor/rules/` or `.cursorrules` | Cursor |
| `windsurf` | `.windsurfrules` | Windsurf |
| `generic` | `AGENTS.md` | Any AI tool |

### Claude Desktop (MCP)

Connect your council via Model Context Protocol.

**1. Configure Claude Desktop**

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "council": {
      "command": "council",
      "args": ["mcp"]
    }
  }
}
```

**2. Restart Claude Desktop**

Claude can now list experts, get details, and consult your council.

### ChatGPT / Gemini / Other AI

Export your council as portable markdown:

```bash
council export              # Output to stdout
council export | pbcopy     # Copy to clipboard (macOS)
council export > council.md # Save to file
```

Use in Custom GPT instructions, Gemini custom instructions, or any AI chat.

---

## Command Reference

### `council init`

Creates the `.council/` directory structure.

```bash
council init
```

Creates:
```
.council/
├── config.yaml     # Configuration
├── experts/        # Expert persona files
└── commands/       # Slash commands
```

### `council detect`

Analyzes your project to detect languages, frameworks, and patterns.

```bash
council detect          # Human-readable output
council detect --json   # JSON output
```

### `council setup`

Generates expert suggestions. See [Setup Modes](#setup-modes) for details.

### `council list`

Shows all experts in your council.

```bash
council list
```

### `council show <id>`

Displays full details of an expert.

```bash
council show dhh
council show kent-beck
```

### `council add`

Adds an expert manually.

```bash
council add "Sandi Metz" --focus "Object design, SOLID principles"
```

### `council remove`

Removes an expert from the council.

```bash
council remove sandi-metz
```

### `council sync`

Syncs your council to AI tool configs. See [Integrations](#integrations).

### `council export`

Exports council as portable markdown. See [ChatGPT / Gemini / Other AI](#chatgpt--gemini--other-ai).

### `council mcp`

Starts an MCP server. See [Claude Desktop (MCP)](#claude-desktop-mcp).

### `council doctor`

Validates your council configuration and reports issues.

```bash
council doctor
```

Checks directory structure, config syntax, expert files, and sync targets.

---

## Configuration

```yaml
# .council/config.yaml
version: 1

# AI CLI for setup --apply
ai:
  command: "claude"    # or "aichat", "llm", "sgpt"
  timeout: 120         # seconds

# Sync targets
targets:
  - claude
  - cursor

# Command generation
council:
  include_council_command: true    # Generate /council command
  include_expert_commands: true    # Generate individual expert commands
```

## Expert Format

Experts are markdown files with YAML frontmatter:

```markdown
---
id: dhh
name: "David Heinemeier Hansson"
focus: "Rails doctrine, conventions, architecture"
triggers:
  paths:
    - "config/routes.rb"
    - "app/controllers/**"
  keywords:
    - "architecture"
    - "service object"
philosophy: |
  Convention over configuration is not just a technical choice.
  It represents a philosophy of trading flexibility for productivity.
principles:
  - "RESTful routes only"
  - "Fat models, skinny controllers"
red_flags:
  - "Service objects everywhere"
  - "Abstract base classes"
---

# DHH - Rails Doctrine

You are channeling David Heinemeier Hansson...
```

## License

MIT
