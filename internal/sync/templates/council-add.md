# Add Expert to Council

Add a new expert to the council: $ARGUMENTS

## Instructions

You are creating a rich expert profile for the council. Parse the arguments to extract:
- **Name**: The expert's name (e.g., "Kent Beck", "Sandi Metz")
- **Focus** (optional): Their area of expertise after `--focus` flag

## Step 1: Check Current Council & Suggest Alternatives

First, run this command to see which experts are already installed:
```bash
council list --json
```

If the user didn't provide a `--focus` flag:
1. Based on the requested name/persona, infer what kind of expertise they're looking for
2. Check if there are similar experts already in the council (avoid duplicates)
3. Suggest 2-3 alternative experts that would complement the council well, based on:
   - The project's tech stack (check package.json, go.mod, etc.)
   - Gaps in current council expertise
   - Well-known experts in relevant domains
4. Present suggestions to the user and let them decide

Example suggestions format:
> Based on your request for "{name}", here are some options:
> 1. **{Name A}** - {their known focus area}
> 2. **{Name B}** - {their known focus area}
> 3. **Custom** - Create a persona with specific focus
>
> Which would you like to add?

## Step 2: Generate Expert Profile

Once the expert is confirmed, research or use your knowledge of this person to generate:

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
2. Run `council sync` to update AI tool configurations
3. Confirm creation with: "Added {Name} ({id}) to the council"
4. Show the file path

Do NOT run the `council add` CLI command - write the file directly with rich content.
