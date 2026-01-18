# Detect Stack and Suggest Experts

Analyze this codebase and suggest council experts.

## Instructions

You have context of the codebase. Analyze it to understand:

1. **Languages**: What programming languages are used? What's the primary language?
2. **Frameworks**: What frameworks, libraries, or tools are in use?
3. **Patterns**: What architectural patterns do you see? (MVC, microservices, monolith, etc.)
4. **Testing**: What testing approaches are used?
5. **Domain**: What problem domain does this project address?

## Suggest Experts

Based on your analysis, suggest 3-5 experts who would be valuable council members for this specific codebase. Consider:

- Language-specific experts (e.g., Rob Pike for Go, Matz for Ruby)
- Framework experts (e.g., DHH for Rails, Chris McCord for Phoenix)
- Pattern/practice experts (e.g., Kent Beck for TDD, Sandi Metz for OO design)
- Domain experts if relevant

For each suggested expert, provide:
- **Name**: The expert's name
- **Focus**: Their specific focus area relevant to THIS project
- **Why**: One sentence on why they'd be valuable for this codebase

## Output Format

Present your findings:

```
## Detected Stack

**Primary Language**: {language}
**Frameworks**: {list}
**Testing**: {tools/approaches}
**Patterns**: {observed patterns}

## Suggested Council

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
