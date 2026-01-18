# Add Expert to Council

Add a new expert to the council: $ARGUMENTS

## Instructions

You are creating a rich expert profile for the council. Parse the arguments to extract:
- **Name**: The expert's name (e.g., "Kent Beck", "Sandi Metz")
- **Focus**: Their area of expertise after --focus flag

If the focus is not provided, ask the user for it.

## Generate Expert Profile

Research or use your knowledge of this person to generate:

1. **Philosophy** (2-4 sentences): What they believe about software/design. Write in first person as if they're speaking. Capture their distinctive worldview.

2. **Principles** (4-6 items): Concrete, actionable guidelines they're known for. These should be memorable and specific to their thinking.

3. **Red Flags** (3-5 items): Patterns they would call out during code review. Things that violate their principles.

## Output Format

Create the expert file at `.council/experts/{id}.md` with this structure:

```markdown
---
id: {kebab-case-id}
name: {Full Name}
focus: {focus area}
philosophy: |
  {philosophy text - first person, 2-4 sentences}
principles:
  - {principle 1}
  - {principle 2}
  - {principle 3}
  - {principle 4}
red_flags:
  - {red flag 1}
  - {red flag 2}
  - {red flag 3}
---

# {Name} - {focus}

You are channeling {Name}, known for expertise in {focus}.

## Philosophy

{philosophy text}

## Principles

- {principle 1}
- {principle 2}
- ...

## Red Flags

Watch for these patterns:
- {red flag 1}
- {red flag 2}
- ...

## Review Style

{2-3 sentences describing how they approach code review}
```

## After Creating

1. Write the file using your file writing capability
2. Confirm creation with: "Added {Name} ({id}) to the council"
3. Show the file path

Do NOT run the `council add` CLI command - write the file directly with rich content.
