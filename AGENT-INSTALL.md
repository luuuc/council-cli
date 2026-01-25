# Install Council CLI

You are an AI assistant helping a user install council-cli. Follow these steps.

**Tool-Specific Notes:**
- **Claude Code**: Use `AskUserQuestion` to present choices as clickable options where indicated. Keep labels short (2-4 words), put details in descriptions.
- **OpenCode**: Present choices as numbered lists. Wait for user to type a number.

## Tell the User What's Happening

Briefly explain what you're about to do, then proceed (no confirmation needed - they asked for setup):

> I'll set up council-cli: install the CLI, create your council with experts matched to your project, and sync to your AI tool. Here we go!

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
| "All set" | Keep everything as is |

- If **Add more**: Skip to "Customize" (Step 4)
- If **Start fresh**: Run `council init --clean`, then run `council start`
- If **All set**: Skip to "Done"

### If .council/ exists but council list fails:

A `.council/` directory exists from a previous or incompatible installation. Use **AskUserQuestion**:

| Label | Description |
|-------|-------------|
| "Start fresh" | Remove and reinitialize |
| "Cancel" | Stop here |

- If **Start fresh**: Run `council init --clean`, then continue to Step 3
- If **Cancel**: Stop here

### If no .council/ directory:

Continue to Step 3.

## Step 3: Set Up Your Council

Run the zero-config setup:

```bash
council start
```

This single command:
1. Creates the `.council/` directory
2. Detects your AI tool (Claude Code, OpenCode, or generic)
3. Detects your project stack (languages, frameworks, testing tools)
4. Adds 5 experts matched to your stack
5. Syncs everything to your AI tool

Output looks like:
```
✓ Detected: Claude Code
✓ Detected: Go, CLI tool
✓ Added 5 experts: Rob Pike, Kent Beck, Jason Fried, Dieter Rams, Sandi Metz

Your council is ready. Try: /council <topic>
```

**That's it.** Your council is ready to use.

## Step 4: Customize (Optional)

Ask if the user wants to customize their council using **AskUserQuestion**:

| Label | Description |
|-------|-------------|
| "Looks good" | Keep the auto-selected experts |
| "Customize" | Add or remove experts |

If **Looks good**: Skip to "Done".

If **Customize**: Help them modify the council.

### Adding Experts

Browse available personas:
```bash
council personas --json
```

Add from the curated library:
```bash
council add "Kent Beck"
council add "Sandi Metz"
```

For experts not in the library, `council add` triggers a creation flow:
```bash
council add "My CTO"
# Prompts for focus and philosophy
```

Or use the `/council-add` skill to search and discover:
```
/council-add a security expert
/council-add someone for API design
```

### Removing Experts

List current experts:
```bash
council list
```

Remove by ID:
```bash
council remove kent-beck
```

### Syncing Changes

After adding or removing experts, sync to update your AI tool:
```bash
council sync
```

## Done

Tell the user setup is complete and list their experts, then use **AskUserQuestion**:

| Label | Description |
|-------|-------------|
| "Try it now" | Run /council on a file or topic |
| "I'm all set" | End the setup flow |

If **Try it now**: Ask what they'd like the council to review (a file, function, or topic), then run `/council` for them.

Remind them of available commands:
- `/council <topic>` - Get expert code reviews
- `/council-add <description>` - Search and add experts by description
