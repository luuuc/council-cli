# Install Council CLI

You are an AI assistant helping a user install council-cli. Follow these steps:

## First: Tell the User What's Happening

Before running any commands, explain to the user:

---

I'll help you set up council-cli. Here's what will happen:

1. **Install** the `council` CLI to ~/.local/bin
2. **Initialize** a `.council/` directory in your project
3. **Detect** your tech stack and suggest relevant experts
4. **Sync** expert profiles to your AI tool config

This takes about a minute. Ready?

---

Wait for confirmation before proceeding.

## Step 1: Install the CLI

Run this command to install council-cli:

```bash
curl -fsSL https://raw.githubusercontent.com/luuuc/council-cli/main/install.sh | sh
```

If that fails (permissions, curl not available, etc.), try:

```bash
go install github.com/luuuc/council-cli/cmd/council@latest
```

## Step 2: Initialize

```bash
council init
```

This creates the `.council/` directory structure.

## Step 3: Detect Stack and Add Experts

Use `/council-detect` to analyze the codebase and suggest experts tailored to this project.

If `/council-detect` is not available yet (first install), run:

```bash
council setup -i
```

This provides interactive expert selection based on detected tech stack.

## Step 4: Sync to Your AI Tool

```bash
council sync
```

This generates slash commands for your AI tool:
- `/council` - Convene the full council for code review
- `/council-add` - Add experts with AI-generated content
- `/council-detect` - Analyze codebase and suggest experts

## Done

Tell the user:
- Installation complete
- They can now use `/council <topic>` to get expert code reviews
- They can add more experts with `/council-add "Name" --focus "area"`
