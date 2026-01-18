# Detect Stack and Suggest Experts

Analyze this codebase and suggest council experts.

## Step 1: Run Detection

First, run the detection command to get structured stack information:

```bash
council detect --json
```

This gives you the detected languages, frameworks, testing tools, and patterns.

## Step 2: Analyze Results

Review the detection output and supplement with your codebase knowledge:

1. **Languages**: What's the primary language by percentage?
2. **Frameworks**: What frameworks are detected?
3. **Testing**: What testing tools/approaches are in use?
4. **Patterns**: What architectural patterns are detected?
5. **Domain**: What problem domain does this project address? (Use your codebase context)

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

```
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
```

## After Analysis

Ask the user which experts they want to add. For each one they choose, use `/council-add {Name} --focus "{focus}"` to create them with rich AI-generated content.

If the user says "all" or "add them", add all suggested experts.
