# Install Council CLI

You are an AI assistant helping a user install council-cli. Follow these steps:

## First: Check Existing Setup

Before doing anything, check if council is already installed:

```bash
council --version 2>/dev/null && council list 2>/dev/null
```

### If council is already installed with experts:

Tell the user:

---

Council is already set up in this project with the following experts:
{list the experts from council list output}

Would you like to:
1. **Keep existing** - Add more experts to your current council
2. **Start fresh** - Remove existing council and set up from scratch
3. **Cancel** - Keep everything as is

---

- If they choose **Keep existing**: Skip to "Add Experts"
- If they choose **Start fresh**: Run `council init --clean`, then continue with setup
- If they choose **Cancel**: Stop here

### If council is installed but no experts:

Skip to "Add Experts" step.

### If council is not installed:

Continue with the full installation flow below.

---

## Tell the User What's Happening

Before running install commands, explain:

---

I'll help you set up council-cli. Here's what will happen:

1. **Install** the `council` CLI to ~/.local/bin
2. **Initialize** a `.council/` directory in your project
3. **Detect** your tech stack and help you choose relevant experts
4. **Sync** expert profiles to your AI tool config

Ready?

---

Wait for confirmation before proceeding.

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

If it says ".council/ already exists", the user can either:
- Use `council init --clean` to remove existing setup and start fresh
- Skip this step and proceed to add experts

## Step 3: Detect Stack

Run detection to understand the project:

```bash
council detect --json
```

This outputs the detected languages, frameworks, testing tools, and patterns.

## Step 4: Add Experts

Based on the detection results, suggest **3-5 experts** (maximum 7) to the user. Consider:

- **Framework experts** (1-2): DHH for Rails, Chris McCord for Phoenix, etc.
- **Language experts** (1-2): Rob Pike for Go, Matz for Ruby, José Valim for Elixir, etc.
- **Practice experts** (1-2): Kent Beck for TDD, Sandi Metz for OO design, etc.

Present your suggestions:

---

Based on your {language} project using {frameworks}, I suggest these experts:

1. **{Name}** - {Focus} - {Why relevant to this project}
2. **{Name}** - {Focus} - {Why relevant to this project}
3. **{Name}** - {Focus} - {Why relevant to this project}

Which experts would you like to add? (all / numbers / none)

---

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

Tell the user:

---

Setup complete! Your council has {N} experts: {names}.

You can now use:
- `/council <topic>` - Get expert code reviews
- `/council-add "Name" --focus "area"` - Add more experts
- `/council-detect` - Re-analyze and get new suggestions

---
