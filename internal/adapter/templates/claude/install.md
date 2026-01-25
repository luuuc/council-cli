# Install Council

Your AI tool will read the appropriate instructions for setting up the council.

## Quick Start

1. Check if council is already set up:
```bash
council list
```

2. If not set up, run:
```bash
council start
```

This single command creates your council, detects your project stack, adds matched experts, and syncs to your AI tool.

## Customization

After setup, you can modify your council:

- `council add "Expert Name"` - add a curated expert
- `council add "Custom Name"` - create a custom expert (if not in library)
- `/council-add` - interactive expert search
- `council remove <id>` - remove an expert
- `council sync` - sync changes to your AI tool
