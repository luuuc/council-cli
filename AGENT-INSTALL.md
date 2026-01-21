# Install Council CLI

You are an AI assistant helping a user install council-cli. Follow these steps.

**Note for Claude Code**: Use `AskUserQuestion` to present choices as clickable options where indicated. Keep labels short (2-4 words), put details in descriptions.

## First: Check Existing Setup

Before doing anything, check the current state:

```bash
council version 2>/dev/null && council list 2>/dev/null
ls -la .council/ 2>/dev/null
```

### If council is installed with experts:

Tell the user which experts exist, then use **AskUserQuestion** with these options:

| Label | Description |
|-------|-------------|
| "Add more" | Keep current experts and add new ones |
| "Start fresh" | Remove everything and set up from scratch |
| "Cancel" | Keep everything as is |

- If **Add more**: Skip to "Add Experts"
- If **Start fresh**: Run `council init --clean`, then continue with setup
- If **Cancel**: Stop here

### If council is installed but no experts:

Skip to "Add Experts" step.

### If council is not installed but .council/ directory exists:

Tell the user a `.council/` directory exists from a previous installation, then use **AskUserQuestion**:

| Label | Description |
|-------|-------------|
| "Keep it" | Install CLI and keep current experts |
| "Start fresh" | Install CLI and set up from scratch |
| "Cancel" | Stop here |

- If **Keep it**: Install CLI, skip init, go to "Sync to Your AI Tool"
- If **Start fresh**: Install CLI, then run `council init --clean`
- If **Cancel**: Stop here

### If council is installed but no .council/ directory:

The CLI is installed but this project hasn't been initialized yet. Skip Step 1 (Install) and continue with Step 2 (Initialize).

### If council is not installed and no .council/ directory:

Continue with the full installation flow below.

---

## Tell the User What's Happening

Briefly explain what you're about to do, then proceed (no confirmation needed - they asked for setup):

> I'll set up council-cli: install the CLI, detect your stack, help you pick experts, and sync to your AI tool. Here we go!

## Step 1: Install the CLI

```bash
curl -fsSL https://raw.githubusercontent.com/luuuc/council-cli/main/install.sh | sh
```

If that fails (permissions, curl not available, etc.), try:

```bash
go install github.com/luuuc/council-cli/cmd/council@latest
```

Note: The install script will skip download if the latest version is already installed.

## Step 2: Initialize

```bash
council init
```

This creates the `.council/` directory structure.

**If it fails with ".council/ already exists"**: Use **AskUserQuestion**:

| Label | Description |
|-------|-------------|
| "Start fresh" | Remove existing council and start over |
| "Keep it" | Skip init and add to existing council |

- If **Start fresh**: Run `council init --clean`
- If **Keep it**: Continue to next step

## Step 3: Detect Stack

Run detection to understand the project:

```bash
council detect --json
```

This outputs the detected languages, frameworks, testing tools, and patterns.

## Step 4: Add Experts

First, check available curated personas that match the detected stack:

```bash
council personas --json
```

This returns experts organized by category (languages, frameworks, practices, etc.) with pre-written philosophies, principles, and red flags. **Prefer suggesting from this list** as these personas are fully fleshed out.

Based on the detection results and available personas, suggest **3-5 experts** (maximum 7) to the user. Mix technical and non-technical perspectives:

**Always suggest (based on detected stack):**
- **Framework expert** (1): DHH for Rails, Chris McCord for Phoenix, etc.
- **Language expert** (1): Rob Pike for Go, Matz for Ruby, José Valim for Elixir, etc.
- **Practice expert** (1): Kent Beck for TDD, Sandi Metz for OO design, etc.
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

```bash
council add "{Name}" --focus "{Focus}"
```

Example:
```bash
council add "Rob Pike" --focus "Go simplicity and idiomatic code"
council add "Kent Beck" --focus "TDD and incremental design"
```

## Step 5: Sync to Your AI Tool

```bash
council sync
```

This generates slash commands for your AI tool:
- `/council` - Convene the full council for code review
- `/council-add` - Add experts with AI-generated content
- `/council-detect` - Analyze codebase and suggest experts

## Done

Tell the user setup is complete and list their experts, then use **AskUserQuestion**:

| Label | Description |
|-------|-------------|
| "Try it now" | Run /council on a file or topic |
| "I'm all set" | End the setup flow |

If **Try it now**: Ask what they'd like the council to review (a file, function, or topic), then run `/council` for them.

Remind them of available commands:
- `/council <topic>` - Get expert code reviews
- `/council-add "Name" --focus "area"` - Add more experts
- `/council-detect` - Re-analyze and get new suggestions
