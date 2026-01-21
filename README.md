# council-cli

AI-agnostic expert council setup for coding assistants.

## Get Started

**In Claude Code or OpenCode, just say:**

> Grab https://raw.githubusercontent.com/luuuc/council-cli/main/AGENT-INSTALL.md and get me set up

That's it. Your AI handles the rest.

## What is a Council?

The council pattern establishes **expert personas** that represent excellence in specific domains. Instead of generic AI review, you get specific expertise in the domains that matter for your project.

```
Your Project           council setup              AI Config Files
┌────────────┐         ┌──────────┐              ┌────────────┐
│ Rails 8    │         │          │              │ .claude/   │
│ Hotwire    │────────▶│ suggests │─────────────▶│ .opencode/ │
│ Minitest   │         │ experts  │              │ AGENTS.md  │
└────────────┘         └──────────┘              └────────────┘
```

After setup, use `/council <topic>` to convene your experts for code review.

## Manual Installation

For contributors or power users who prefer manual setup:

```bash
# Direct download (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/luuuc/council-cli/main/install.sh | sh

# Go
go install github.com/luuuc/council-cli/cmd/council@latest

# From source (see CONTRIBUTING.md)
git clone https://github.com/luuuc/council-cli.git && cd council-cli && make install
```

Then initialize and set up your council:

```bash
council init
council setup -i      # Interactive mode
council sync
```

## Commands

| Command | Description |
|---------|-------------|
| `council init` | Initialize `.council/` directory |
| `council setup -i` | Interactive expert selection |
| `council setup --apply` | AI-assisted expert suggestions |
| `council detect` | Detect project stack |
| `council list` | List your experts |
| `council add "Name" --focus "area"` | Add an expert |
| `council remove <id>` | Remove an expert |
| `council sync` | Sync to AI tool configs |
| `council export` | Export as portable markdown |
| `council doctor` | Validate configuration |

## Configuration

```yaml
# .council/config.yaml
version: 1

ai:
  command: "claude"
  timeout: 120

targets:
  - claude
  - opencode
  - generic
```

## Other AI Tools

For Cursor, Windsurf, ChatGPT, Gemini, or any other AI:

1. Set up council via Claude Code, OpenCode, or manual installation
2. Run `council export | pbcopy`
3. Paste into your AI tool's custom instructions

<details>
<summary>Legacy targets (Cursor/Windsurf)</summary>

If you have an existing config with `cursor` or `windsurf` targets, they still work:

```bash
council sync cursor
council sync windsurf
```

For new projects, we recommend using `council export` instead.

</details>

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

MIT
