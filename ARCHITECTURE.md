# Architecture

```
council start / council add "Name"
         │
         ▼
┌─────────────────┐
│  detect.Scan()  │  Analyzes project files (go.mod, Gemfile, package.json)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  expert.Save()  │  Writes to .council/experts/*.md (YAML frontmatter + markdown)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  sync.SyncAll() │  Reads experts, determines target adapter
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Adapter      │  Formats and writes to tool-specific locations
└─────────────────┘
         │
         ├── Claude:   .claude/agents/*.md, .claude/commands/*.md
         ├── OpenCode: .opencode/agents/*.md, .opencode/commands/*.md
         └── Generic:  AGENTS.md
```

## Key Packages

| Package | Purpose |
|---------|---------|
| `detect` | Identifies languages, frameworks, testing tools from project files |
| `expert` | Expert data model: parse, save, list markdown files with YAML frontmatter |
| `adapter` | Tool-specific formatting and paths (Claude, OpenCode, Generic) |
| `sync` | Orchestrates writing experts to AI tool configurations |
| `config` | Manages `.council/config.yaml` |
| `install` | Manages external persona repositories (git clone/pull) |

## Adapter Pattern

Each AI tool implements the `Adapter` interface:

```go
type Adapter interface {
    Name() string           // "claude", "opencode", "generic"
    Detect() bool           // Does this tool exist in current project?
    Paths() Paths           // Where to write files
    FormatAgent(*Expert)    // Tool-specific expert format
    FormatCommand(...)      // Tool-specific command format
}
```

Adding a new tool: implement `Adapter`, call `Register()` in `init()`.

## Expert File Format

```markdown
---
id: kent-beck
name: Kent Beck
focus: Test-driven development
philosophy: |
  Write tests first.
principles:
  - Red-green-refactor
red_flags:
  - Tests written after code
---

# Kent Beck - Test-driven development

Body content here...
```

## Data Flow

1. **Detection**: `detect.Scan()` walks project, counts file extensions, checks config files
2. **Storage**: Experts saved to `.council/experts/` as markdown with YAML frontmatter
3. **Sync**: `sync.SyncAll()` loads experts, selects adapter, writes to tool directories
4. **Commands**: Synced commands (`/council`, `/council-add`) invoke experts during AI sessions
