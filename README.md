# council-cli

AI-agnostic expert council setup for coding assistants.

Tired of generic AI code reviews that feel like talking to a wall? Getting the same bland suggestions over and over? Want your AI to actually challenge you like a real mentor would?

**Council CLI is here for you.** Build your own expert council and make AI code review feel sharp, opinionated, and actually useful again.

## The Council Pattern

Instead of generic AI code review, establish **expert personas** that represent excellence in specific domains:

```
Your Project           council setup              AI Config Files
┌────────────┐         ┌──────────┐              ┌────────────┐
│ Rails 8    │         │          │              │ .claude/   │
│ Hotwire    │────────▶│    AI    │─────────────▶│ .cursor/   │
│ Minitest   │         │ suggests │              │ .windsurf/ │
│ SQLite     │         │ experts  │              │ AGENTS.md  │
└────────────┘         └──────────┘              └────────────┘
```

The AI analyzes your tech stack and suggests appropriate experts. You review, approve, and the CLI syncs to your AI tools.

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

## Quick Start

```bash
# Initialize council directory
council init

# Analyze project and get AI prompt for expert suggestions
council setup

# Or let AI suggest and apply directly (requires AI CLI like claude)
council setup --apply

# Sync to AI tool configs
council sync
```

## Commands

### `council init`

Creates the `.council/` directory structure in your project.

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

Analyzes your project to detect languages, frameworks, testing tools, and patterns.

```bash
council detect          # Human-readable output
council detect --json   # JSON output
```

### `council setup`

Generates an AI prompt for expert suggestions based on your project analysis.

```bash
council setup                    # Output prompt to copy to AI
council setup -o prompt.md       # Save prompt to file
council setup --apply            # Send to configured AI CLI
council setup --apply --yes      # Skip confirmation
```

You can also pipe AI responses:
```bash
council setup --apply < response.yaml
```

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

Syncs your council to AI tool configs.

```bash
council sync                # Sync to all configured targets
council sync claude         # Sync to specific target
council sync --dry-run      # Preview changes
```

### `council export`

Exports your council as portable markdown for use anywhere.

```bash
council export              # Output to stdout
council export | pbcopy     # Copy to clipboard (macOS)
council export | xclip      # Copy to clipboard (Linux)
council export > council.md # Save to file
```

Use exported markdown in:
- **Desktop AI apps** - Claude Desktop, ChatGPT, Gemini custom instructions
- **Web interfaces** - Paste into any AI chat
- **Sharing** - Send council to a colleague

### `council mcp`

Starts an MCP server for Claude Desktop integration.

```bash
council mcp
```

See [Use with Claude Desktop](#use-with-claude-desktop) for setup instructions.

## Configuration

```yaml
# .council/config.yaml
version: 1

# AI CLI to use for setup --apply
ai:
  command: "claude"    # or "aichat", "llm", "sgpt"
  timeout: 120         # seconds

# Which AI tools to sync to
targets:
  - claude
  - cursor

# Command generation options
council:
  include_council_command: true    # Generate /council command
  include_expert_commands: true    # Generate individual expert commands
```

## Sync Targets

| Target | Location | Description |
|--------|----------|-------------|
| `claude` | `.claude/agents/`, `.claude/commands/` | Claude Code |
| `cursor` | `.cursor/rules/` or `.cursorrules` | Cursor |
| `windsurf` | `.windsurfrules` | Windsurf |
| `generic` | `AGENTS.md` | Any AI tool |

## Expert Format

Experts are stored as markdown files with YAML frontmatter:

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

## Integrations

### Use with Claude Desktop

Connect your council to Claude Desktop via MCP (Model Context Protocol).

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

Your council is now available. Claude can:
- List your experts with `list_experts`
- Get expert details with `get_expert`
- Consult all experts with `consult_council`
- Use the `/council` prompt template

**Example conversation:**

```
You: Review this code with my council

Claude: [Uses consult_council tool]
        Let me consult your expert council...

        DHH says: The service object here is unnecessary...
        Kent Beck says: I don't see tests for this...
        Sandi Metz says: This class has too many responsibilities...
```

### Use with ChatGPT

Create a Custom GPT with your council's expertise.

**1. Export your council**

```bash
council export | pbcopy  # macOS
council export | xclip   # Linux
```

**2. Create a Custom GPT**

1. Go to [ChatGPT](https://chat.openai.com) → Explore GPTs → Create
2. Name it "My Council" (or any name you prefer)
3. In Instructions, paste:

```markdown
You have access to an expert council. When reviewing work, consult these experts:

[PASTE YOUR COUNCIL EXPORT HERE]

When the user asks for a review:
1. Consider each expert's perspective
2. Identify issues each would raise
3. Synthesize into actionable feedback
```

4. Save and publish (private or public)

**3. Use your GPT**

Start a chat with your Custom GPT and ask it to review code, architecture decisions, or any work relevant to your council's expertise.

### Use with Other AI Tools

The `council export` command outputs portable markdown that works anywhere:

- **Gemini** - Paste into custom instructions
- **Perplexity** - Start conversations with council context
- **Local LLMs** - Include in system prompts
- **API integrations** - Use as context in your applications

## Why This Pattern?

The council pattern:
- **Defines standards** through expert personas
- **Makes them autonomous** with clear scope and strong opinions
- **Introduces structure** via commands like `/council`, `/dhh`, `/kent`

Instead of starting with generic AI review, you start with specific expertise in the domains that matter for your project.

## License

MIT
