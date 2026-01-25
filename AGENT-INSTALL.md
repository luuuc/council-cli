# Install Council CLI

You are an AI assistant helping a user install council-cli. Follow these steps.

**Tool-Specific Notes:**
- **Claude Code**: Use `AskUserQuestion` to present choices as clickable options where indicated. Keep labels short (2-4 words), put details in descriptions.
- **OpenCode**: Present choices as numbered lists. Wait for user to type a number.

## Tell the User What's Happening

Briefly explain what you're about to do, then proceed (no confirmation needed - they asked for setup):

> I'll set up council-cli: install/update the CLI, detect your stack, help you pick experts, and sync to your AI tool. Here we go!

## Step 1: Install or Update the CLI

**Always run this first**, regardless of whether council is already installed. The script is idempotent: it installs if missing, updates if outdated, and does nothing if current.

```bash
curl -fsSL https://raw.githubusercontent.com/luuuc/council-cli/main/install.sh | sh
```

If that fails (permissions, curl not available, etc.), try:

```bash
go install github.com/luuuc/council-cli/cmd/council@latest
```

## Step 2: Check Existing Setup

After ensuring the CLI is current, check for existing configuration:

```bash
council list 2>/dev/null
```

### If experts already exist:

Tell the user which experts exist, then use **AskUserQuestion**:

| Label | Description |
|-------|-------------|
| "Add more" | Keep current experts and add new ones |
| "Start fresh" | Remove everything and set up from scratch |
| "Cancel" | Keep everything as is |

- If **Add more**: Skip to "Add Experts" (Step 5)
- If **Start fresh**: Run `council init --clean`, then continue with setup
- If **Cancel**: Stop here

### If .council/ exists but no experts:

Continue to "Detect Stack" (Step 4).

### If .council/ directory exists but council list returns non-zero exit code:

A `.council/` directory exists from a previous or incompatible installation. Use **AskUserQuestion**:

| Label | Description |
|-------|-------------|
| "Start fresh" | Remove and reinitialize |
| "Cancel" | Stop here |

- If **Start fresh**: Run `council init --clean`, then continue
- If **Cancel**: Stop here

### If no .council/ directory:

Continue to "Initialize" (Step 3).

## Step 3: Initialize

```bash
council init
```

This creates the `.council/` directory structure and auto-detects your AI tool (Claude Code, OpenCode, or generic).

If multiple tools are detected, it will prompt you to choose. You can also specify explicitly:
```bash
council init --tool=claude
council init --tool=opencode
council init --tool=generic
```

**Configure yourself as the AI command:**

After init, edit `.council/config.yaml` and set your CLI as the AI command:

```yaml
version: 1
ai:
  command: claude  # Set this to your own CLI: claude, opencode, etc.
  timeout: 120
```

This ensures `council setup --apply` uses you for AI operations.

**Run initial sync to enable slash commands:**

```bash
council sync
```

This makes `/council`, `/council-add`, and `/council-detect` available immediately, so you can use them during this setup session without needing to restart.

## Step 4: Detect Stack

Run detection to understand the project:

```bash
council detect --json
```

This outputs the detected languages, frameworks, testing tools, and patterns.

### If Detection Is Sparse

If `council detect --json` returns no frameworks and no patterns (only languages or empty), the project context is unclear. Ask the user what they're building using **AskUserQuestion**:

| Label | Description |
|-------|-------------|
| "Code quality" | Code review, refactoring, testing, architecture |
| "Product/UX" | User-facing features, design decisions, frontend work |
| "Ops/Process" | DevOps, CI/CD, team workflow, infrastructure |
| "Business" | Strategy, growth, monetization, leadership |

Use this to weight expert suggestions:

- **Code quality**: Prioritize practice experts (Kent Beck, Sandi Metz, Martin Fowler)
- **Product/UX**: Include design experts (Dieter Rams, Steve Schoger, Jason Fried)
- **Ops/Process**: Include ops experts (Gene Kim, Charity Majors, Kelsey Hightower)
- **Business**: Include business experts (Arvid Kahl, Rob Walling, Sahil Lavingia)

Continue to Step 5 with this context in mind.

## Step 5: Add Experts

First, check available curated personas that match the detected stack:

```bash
council personas --json
```

This returns experts organized by category (languages, frameworks, practices, etc.) with pre-written philosophies, principles, and red flags. **Prefer suggesting from this list** as these personas are fully fleshed out.

Based on the detection results and available personas, suggest **5 experts** (maximum 7) to the user. Mix technical and non-technical perspectives:

**Always suggest (based on detected stack):**
- **Framework expert** (1): DHH for Rails, Chris McCord for Phoenix, etc.
- **Language expert** (1): Rob Pike for Go, Matz for Ruby, José Valim for Elixir, etc.
- **Practice expert** (2): Kent Beck for TDD, Sandi Metz for OO design, etc.
- **General** (1, cross-cutting principles): Cal Newport, Dieter Rams, Gene Kim, etc.

**Suggest based on project context** (ask user if unclear):
- **Product** (if building user-facing features): Marty Cagan, Jason Fried, Teresa Torres
- **Process** (if team/workflow questions): Ryan Singer, Allen Holub, Gene Kim
- **Business** (if startup/monetization): Arvid Kahl, Rob Walling, Sahil Lavingia
- **Design/UX** (if UI work): Dieter Rams, Steve Schoger, Maggie Appleton
- **Growth** (if scaling): Andrew Chen, Brian Balfour, Hiten Shah
- **Leadership** (if managing people): Camille Fournier, Will Larson, Lara Hogan
- **Security** (if auth/data): Troy Hunt, Scott Helme, Tanya Janca
- **Data** (if analytics/ML): DJ Patil, Hilary Mason, Monica Rogati

Check `council personas --json` for the full list organized by category.

Present your suggestions, then use **AskUserQuestion** (this is the key engagement moment!):

| Label | Description |
|-------|-------------|
| "Add all" | Add all suggested experts to your council |
| "Let me choose" | Pick specific experts from the list |
| "Skip for now" | Continue without adding experts |

If **Let me choose**: Present each expert as a selectable option (use `multiSelect: true`), with the expert's focus as the description.

After selection, ask if they want more suggestions:

| Label | Description |
|-------|-------------|
| "Looks good" | Proceed to sync |
| "Suggest more" | Show additional expert recommendations |

For each expert the user approves, run:

**For curated experts** (from `council personas --json`):
```bash
council add "{Name}"
```

Example:
```bash
council add "Rob Pike"
council add "Kent Beck"
```

**For custom experts** (not in curated library):
Use `/council-add` to search and discover experts:
```
/council-add a testing expert
/council-add someone for API design
```

The skill searches curated personas first, presents 4 options, and generates profiles for non-curated selections.

## Step 6: Sync to Your AI Tool

```bash
council sync
```

This updates the slash commands with the newly added experts. The council member list in `/council` will now include all the experts you added.

## Done

Tell the user setup is complete and list their experts, then use **AskUserQuestion**:

| Label | Description |
|-------|-------------|
| "Try it now" | Run /council on a file or topic |
| "I'm all set" | End the setup flow |

If **Try it now**: Ask what they'd like the council to review (a file, function, or topic), then run `/council` for them.

Remind them of available commands:
- `/council <topic>` - Get expert code reviews
- `/council-add a testing expert` - Search and add experts by description
- `/council-detect` - Re-analyze and get new suggestions
