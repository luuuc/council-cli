# Install Council

Set up the council for your project.

## Quick Start

1. Check if council is already initialized:
```bash
ls -la .council/
```

2. If not initialized, run:
```bash
council init
```

3. Analyze your project and suggest experts:
```bash
council detect --json
```

4. Add experts to your council - use one of:
   - `/council-add` - interactive expert search
   - `council add "Expert Name"` - add specific curated expert
   - `council setup --apply` - AI-assisted setup

5. Sync to your AI tool:
```bash
council sync
```
