# MACHINES.md - AI Agent Integration

aigogo supports AI agent discovery and usage through two mechanisms: structured manifest metadata and a Claude Code skill. These serve different audiences and work at different layers.

## Skills vs AI Metadata

**Skills** (`.claude/commands/aigogo.md`) are for **humans using an AI assistant**. A developer types `/aigogo` and Claude walks them through building or consuming a package interactively. The skill is instructions that shape Claude's behavior -- it's a prompt, not data. The human is in the loop making decisions.

**AI metadata** (`ai` field in `aigogo.json`) is for **agents acting autonomously**. An agent is working on a task, needs an HTTP client with retry logic, and can programmatically search available aigogo packages by reading their `ai.summary` and `ai.capabilities` fields to find a match. It then runs `aigogo add` and `aigogo install` itself, and uses the `ai.usage` field as a template to write the integration code. No human interaction required.

| | Skill | AI Metadata |
|---|---|---|
| **Audience** | Human developer via AI assistant | Autonomous agent |
| **When** | User explicitly invokes `/aigogo` | Agent discovers packages during a task |
| **How** | Prompt instructions shaping assistant behavior | Structured data in the package manifest |
| **Interaction** | Interactive, conversational | Programmatic, no human in the loop |
| **Scope** | Knows the aigogo CLI | Describes one specific snippet |

They complement each other: a skill can use AI metadata when helping a human ("I found 3 packages matching your needs, here's what each does"), and an autonomous agent doesn't need the skill at all -- it just reads manifests and runs commands.

The gap right now is discovery. The `ai` field describes individual packages, but there's no aggregation layer -- no index or search that lets an agent query "show me all packages with capability X" across a registry. That would be the next piece for fully autonomous agent consumption.

## AI Metadata in aigogo.json

The `ai` field in `aigogo.json` is an optional section that describes a snippet in terms AI agents can parse and act on. This allows agents to discover, evaluate, and use snippets without reading the source code.

### Schema

```json
{
  "ai": {
    "summary": "One sentence describing what this snippet does and when to use it.",
    "capabilities": [
      "Short verb phrase for each action the code can perform"
    ],
    "usage": "from aigogo.package_name import func\nresult = func(args)",
    "inputs": "Description of expected inputs and their types",
    "outputs": "Description of return values and their types"
  }
}
```

### Fields

| Field | Required | Purpose |
|---|---|---|
| `summary` | Yes | What the snippet does, in one sentence. Should be specific enough for an agent to decide whether to use it. |
| `capabilities` | Yes | List of actions the code can perform, as short verb phrases. Agents use these to match snippets to tasks. |
| `usage` | No | Minimal import and function call example. Shows the agent exactly how to invoke the code. |
| `inputs` | No | What the functions expect. Types, formats, constraints. |
| `outputs` | No | What the functions return. Types, formats, side effects. |

### Design Principles

**Summary** should answer "when would I use this?" not just "what is this?". Compare:
- Bad: "A logging module."
- Good: "Set up structured JSON logging with per-request context injection, useful for log aggregation services."

**Capabilities** should be action-oriented and specific:
- Bad: "logging", "formatting"
- Good: "Configure a JSON-formatted logger", "Inject context fields into all log entries"

**Usage** should be copy-pasteable. An agent should be able to read this field and immediately write working code that uses the snippet.

### Example

From `examples/tool-use-decorator/aigogo.json`:

```json
{
  "name": "tool-use-decorator",
  "version": "1.0.0",
  "description": "Convert Python functions into LLM tool-calling schemas",
  "language": {
    "name": "python",
    "version": ">=3.8,<4.0"
  },
  "files": {
    "include": ["tools.py"]
  },
  "ai": {
    "summary": "Decorate Python functions to auto-generate OpenAI-compatible tool-calling schemas from type hints and docstrings.",
    "capabilities": [
      "Convert a decorated function into a tool-calling JSON schema",
      "Maintain a registry of all decorated tools",
      "Dispatch tool calls by name with argument parsing",
      "Extract parameter descriptions from Google-style docstrings"
    ],
    "usage": "from aigogo.tool_use_decorator import tool, get_tools, call_tool\n\n@tool\ndef get_weather(city: str) -> str:\n    ...\n\ntools = get_tools()\nresult = call_tool('get_weather', '{\"city\": \"London\"}')",
    "inputs": "Python functions with type hints and docstrings",
    "outputs": "OpenAI-compatible tool schema dicts, tool call dispatch"
  }
}
```

### How Agents Use This

1. **Discovery**: An agent searching for "HTTP client with retry" can match against `summary` and `capabilities` fields across available packages.
2. **Evaluation**: The `inputs`, `outputs`, and `usage` fields let the agent determine whether the snippet fits its needs without reading source.
3. **Integration**: The agent runs `aigogo add <package>` and `aigogo install`, then uses the `usage` example as a template for writing code that calls the snippet.

The `ai` field is ignored by all existing aigogo commands (build, install, push, etc.) -- it is purely advisory metadata for agent consumption.

## Claude Code Skill

The repository includes a Claude Code slash command at `.claude/commands/aigogo.md`. When working in this repository (or any project that references this command), users can invoke `/aigogo` to get AI-assisted help with aigogo workflows.

### What the Skill Does

The `/aigogo` command guides Claude through:

- **Creating packages**: Initializes the manifest, detects language, adds files and dependencies, populates AI metadata, and builds.
- **Consuming packages**: Adds packages from registries or local cache, installs them, and shows the correct import syntax for the project's language.
- **Publishing**: Validates, builds, and pushes with the correct `--from` flag.

### How It Works

The skill is a prompt file that teaches Claude the aigogo command set, the two-step build/push workflow, the manifest schema (including the `ai` field), and the import conventions for each language. It is not executable code -- it is instructions that Claude follows when the user invokes the command.

### Distributing the Skill

The skill file lives at `.claude/commands/aigogo.md` in this repository. To use it in another project:

1. Copy `.claude/commands/aigogo.md` into the target project's `.claude/commands/` directory.
2. Invoke with `/aigogo` in Claude Code.

Alternatively, reference it from the project's `CLAUDE.md` by describing the aigogo workflow and pointing to this repository's documentation.

## Go Type Definition

The `ai` field is defined in `pkg/manifest/types.go`:

```go
type AISpec struct {
    Summary      string   `json:"summary"`
    Capabilities []string `json:"capabilities"`
    Usage        string   `json:"usage,omitempty"`
    Inputs       string   `json:"inputs,omitempty"`
    Outputs      string   `json:"outputs,omitempty"`
}
```

It is an optional field on the `Manifest` struct:

```go
type Manifest struct {
    // ...existing fields...
    AI *AISpec `json:"ai,omitempty"`
}
```

The field serializes to/from JSON automatically. No validation is enforced -- it is advisory metadata. The `omitempty` tag ensures it is not written to `aigogo.json` unless explicitly set.

## Examples

See `examples/` for six complete packages with AI metadata:

| Example | Language | What it demonstrates |
|---|---|---|
| `prompt-templates` | Python | Structured prompt templates with variable substitution and chaining |
| `llm-response-parser` | Python | Extract and validate JSON/lists from raw LLM output |
| `embedding-search` | Python | Vector similarity search with numpy (pyproject.toml dependency import) |
| `tool-use-decorator` | Python | Auto-generate OpenAI tool schemas from decorated functions |
| `agent-context-manager` | Python | Sliding-window conversation manager with token budgets (pyproject.toml dependency import) |
| `token-budget-js` | JavaScript | Token estimation, budget checking, and text chunking (zero runtime dependencies) |

Each example includes a complete `aigogo.json` with the `ai` field populated, ready to build and use.
