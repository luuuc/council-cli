# Integrations Guide

This guide covers how to use your council with different AI tools and platforms.

## Claude Desktop (MCP)

Claude Desktop supports the Model Context Protocol (MCP), allowing local tools to extend Claude's capabilities. The `council mcp` command starts an MCP server that exposes your council to Claude Desktop.

### Setup

1. Ensure council-cli is installed and in your PATH:

```bash
which council  # Should output the path to council
```

2. Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "council": {
      "command": "council",
      "args": ["mcp"]
    }
  }
}
```

3. Restart Claude Desktop completely (quit and reopen)

4. Verify the connection by asking Claude: "What tools do you have available?"

### What's Exposed

The MCP server provides:

| Feature | Description |
|---------|-------------|
| `list_experts` tool | Lists all experts with their focus areas |
| `get_expert` tool | Gets detailed info about a specific expert |
| `consult_council` tool | Gets perspectives from all experts |
| `council://experts/{id}` resource | Direct access to expert profiles |
| `/council` prompt | Template for expert review sessions |

### Usage Examples

**List your council:**
> "Show me my expert council"

**Get specific expert advice:**
> "What would DHH say about this service object pattern?"

**Full council review:**
> "Review this code with my council"
> [paste code]

### Troubleshooting

**Council not appearing:**
- Ensure `council` is in your PATH
- Check Claude Desktop logs: `~/Library/Logs/Claude/`
- Verify your `.council/` directory exists with experts

**"council not initialized" errors:**
- Run `council init` in the directory where you want to use your council
- The MCP server uses the current working directory

## ChatGPT Custom GPT

Create a Custom GPT that embodies your council's expertise.

### Setup

1. Export your council:

```bash
council export | pbcopy  # macOS
# or
council export > council.md  # Save to file
```

2. Go to [ChatGPT](https://chat.openai.com)

3. Click your profile → My GPTs → Create a GPT

4. Configure your GPT:

**Name:** My Expert Council (or your preferred name)

**Description:** A personalized expert council for code review and technical guidance.

**Instructions:**
```
You have access to an expert council. When reviewing work or answering questions, consider each expert's perspective and provide feedback as if channeling their expertise.

## Your Council

[PASTE THE OUTPUT OF `council export` HERE]

## How to Respond

When the user asks for a review or advice:
1. Identify which experts are most relevant to the question
2. Consider each relevant expert's perspective
3. Present feedback organized by expert, using their voice and principles
4. Synthesize into actionable recommendations

When presenting expert perspectives:
- Use first person as if the expert is speaking
- Reference their specific principles and red flags
- Be direct and opinionated, as each expert would be
```

5. Save your GPT (private or link-only sharing recommended for personal councils)

### Usage

Start a conversation with your Custom GPT:

> "Review this Rails controller for me"
> [paste code]

The GPT will respond with perspectives from relevant experts in your council.

### Keeping Your GPT Updated

When you update your local council:

1. Re-export: `council export | pbcopy`
2. Edit your GPT
3. Replace the council section in Instructions
4. Save

## Gemini

Google Gemini supports custom instructions that persist across conversations.

### Setup

1. Export your council:

```bash
council export
```

2. Go to [Gemini](https://gemini.google.com)

3. Click Settings → Gemini Apps → Extensions → Custom Instructions

4. Add your council export to the instructions

### Usage

Your council context will be available in all Gemini conversations. Ask:

> "From my council's perspective, what do you think of this architecture?"

## Local LLMs (Ollama, LM Studio, etc.)

Use your council as a system prompt for local language models.

### Setup

1. Export your council:

```bash
council export > system-prompt.md
```

2. Configure your local LLM to use this as the system prompt

**Ollama example:**

```bash
# Create a Modelfile
cat > Modelfile << 'EOF'
FROM llama3.1
SYSTEM """
You have access to an expert council for code review.

$(cat system-prompt.md)

When asked to review code, consider each expert's perspective.
"""
EOF

ollama create council-reviewer -f Modelfile
ollama run council-reviewer
```

**LM Studio:**
Copy the contents of `system-prompt.md` into the System Prompt field.

## API Integrations

Include your council in API calls to any LLM.

### OpenAI API

```python
import openai

# Load your council
with open('council.md', 'r') as f:
    council = f.read()

response = openai.chat.completions.create(
    model="gpt-4",
    messages=[
        {
            "role": "system",
            "content": f"""You have access to an expert council:

{council}

Review code from each expert's perspective."""
        },
        {
            "role": "user",
            "content": "Review this code: ..."
        }
    ]
)
```

### Anthropic API

```python
import anthropic

with open('council.md', 'r') as f:
    council = f.read()

client = anthropic.Anthropic()
message = client.messages.create(
    model="claude-sonnet-4-20250514",
    max_tokens=1024,
    system=f"""You have access to an expert council:

{council}

Review code from each expert's perspective.""",
    messages=[
        {"role": "user", "content": "Review this code: ..."}
    ]
)
```

## Best Practices

1. **Keep your council focused** - 3-5 experts is usually optimal
2. **Update regularly** - Re-sync when you add or modify experts
3. **Match experts to project** - Your Rails project council differs from your Go project council
4. **Test with real code** - Verify experts give useful, distinct perspectives
