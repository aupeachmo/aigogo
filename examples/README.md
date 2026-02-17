# Examples

Six example packages demonstrating aigogo for AI/LLM development. Five Python packages each include a `pyproject.toml` to show dependency import, plus one JavaScript package. All include an `ai` field in `aigogo.json` for agent discovery.

## 1. Prompt Templates (`prompt-templates/`)

Structured prompt templates with variable substitution and chaining. Includes built-in templates for code review, summarization, structured extraction, and chain-of-thought.

```bash
# Author workflow
cd examples/prompt-templates
aigg init                        # Already has aigogo.json
aigg build prompt-templates:1.0.0

# Consumer workflow
cd ~/my-project
aigg add prompt-templates:1.0.0
aigg install
```

```python
from aigogo.prompt_templates import CODE_REVIEW, CHAIN_OF_THOUGHT, Prompt

# Use a built-in template
messages = CODE_REVIEW.render(language="python", code="def foo(): pass")

# Chain prompts: review code, then refine based on feedback
from aigogo.prompt_templates import REFINE
chain = CODE_REVIEW.chain(REFINE)
followup = chain(first_response)

# Create a custom template
classify = Prompt(
    system="Classify the sentiment of the following text.",
    user="{text}",
)
messages = classify.render(text="I love this product!")
```

## 2. LLM Response Parser (`llm-response-parser/`)

Extract and validate structured data from raw LLM output. Handles markdown code blocks, trailing commas, single quotes, and other common LLM formatting issues.

```bash
cd examples/llm-response-parser
aigg build llm-response-parser:1.0.0
```

```python
from aigogo.llm_response_parser import extract_json, extract_list, validate_keys

# Extract JSON from a response that wraps it in ```json blocks
data = extract_json('Here is the result:\n```json\n{"name": "Alice", "age": 30}\n```')

# Handles LLM mistakes: trailing commas, single quotes
data = extract_json("{'name': 'Bob', 'age': 25,}")

# Extract a numbered or bulleted list
items = extract_list("1. First item\n2. Second item\n3. Third item")

# Validate structure
validated = validate_keys(data, required=["name", "age"], optional=["email"])
```

## 3. Embedding Search (`embedding-search/`)

Vector similarity search for RAG pipelines and semantic search. Depends on `numpy` -- demonstrates `pyproject.toml` dependency import.

```bash
cd examples/embedding-search

# Import dependencies from pyproject.toml
aigg add dep --from-pyproject

# Build
aigg build embedding-search:1.0.0
```

```python
from aigogo.embedding_search import cosine_similarity, top_k, deduplicate

# Compare two embeddings
score = cosine_similarity(query_embedding, doc_embedding)

# Find the 5 most similar documents
corpus = [("doc1", emb1), ("doc2", emb2), ("doc3", emb3)]
results = top_k(query_embedding, corpus, k=5)

# Remove near-duplicates from a collection
unique_docs = deduplicate(corpus, threshold=0.95)
```

## 4. Tool-Use Decorator (`tool-use-decorator/`)

Decorate Python functions to auto-generate OpenAI-compatible tool-calling schemas. Zero dependencies.

```bash
cd examples/tool-use-decorator
aigg build tool-use-decorator:1.0.0
```

```python
from aigogo.tool_use_decorator import tool, get_tools, call_tool

@tool
def get_weather(city: str, units: str = "celsius") -> str:
    """Get the current weather for a city.

    Args:
        city: The city name to look up.
        units: Temperature units (celsius or fahrenheit).
    """
    return f"22 degrees {units} in {city}"

# Pass to OpenAI API
response = client.chat.completions.create(
    model="gpt-4",
    messages=messages,
    tools=get_tools(),
)

# Dispatch the tool call
result = call_tool("get_weather", '{"city": "London"}')
```

## 5. Agent Context Manager (`agent-context-manager/`)

Sliding-window conversation manager that auto-trims to a token budget. Depends on `tiktoken` -- demonstrates `pyproject.toml` dependency import.

```bash
cd examples/agent-context-manager

# Import dependencies from pyproject.toml
aigg add dep --from-pyproject

# Build
aigg build agent-context-manager:1.0.0
```

```python
from aigogo.agent_context_manager import ContextWindow

ctx = ContextWindow(max_tokens=4096, reserve_tokens=512)
ctx.set_system("You are a helpful assistant.")

# Conversation loop
ctx.add_user("What is Python?")
ctx.add_assistant("Python is a programming language...")
ctx.add_user("What about Go?")

# Render auto-trims old messages to fit budget
messages = ctx.render()
print(f"Available tokens for response: {ctx.available_tokens()}")

# Use tiktoken for precise counting
import tiktoken
enc = tiktoken.encoding_for_model("gpt-4")
ctx = ContextWindow(max_tokens=8192, token_counter=lambda s: len(enc.encode(s)))
```

## 6. Token Budget JS (`token-budget-js/`)

Token estimation, budget checking, and text chunking for LLM APIs. Zero runtime dependencies -- demonstrates a JavaScript package with `package.json`.

```bash
cd examples/token-budget-js

# Build
aigg build token-budget-js:1.0.0
```

```javascript
const { countTokens, checkBudget, chunkText, trimConversation } = require('@aigogo/token-budget-js');

// Count tokens in a string
const count = countTokens("Hello, world!");

// Check if messages fit in a context window
const messages = [
  { role: "system", content: "You are a helpful assistant." },
  { role: "user", content: "Explain quantum computing." },
];
const { fits, total, remaining } = checkBudget(messages, 4096);

// Split long text into token-limited chunks for batch processing
const chunks = chunkText(longDocument, 512);

// Trim conversation to fit budget (preserves system prompt)
const trimmed = trimConversation(messages, 4096);
```

## pyproject.toml Workflow

Three of the Python examples (`embedding-search`, `agent-context-manager`, `prompt-templates`) include a `pyproject.toml`. This demonstrates how aigg imports dependencies from existing Python project files:

```bash
# Instead of manually adding each dependency:
aigg add dep numpy ">=1.24.0,<2.0.0"

# Import them all at once from pyproject.toml:
aigg add dep --from-pyproject
```

This is useful when packaging existing code that already has a `pyproject.toml` -- aigg reads the dependency list and adds them to `aigogo.json` automatically.

## Dependencies

aigg manages **AI agents**, not environments. `aigg install` pulls the source files and creates import symlinks, but it does not run `pip install` or `npm install`. Dependencies declared in `aigogo.json` are metadata that tells the consumer what their environment needs.

For packages with dependencies, the consumer installs them separately using their preferred package manager. The `show-deps` command outputs dependencies in various formats to make this easy.

### Python Dependencies

```bash
# Install the snippet
aigg add embedding-search:1.0.0
aigg install
```

**pyproject.toml (PEP 621) — recommended:**
```bash
# Output aigogo-specific optional-dependency groups (copy into your pyproject.toml)
aigg show-deps .aigogo/imports/aigogo/embedding_search --format pyproject

# Then install with:
pip install -e '.[aigogo]'
```

**Poetry:**
```bash
# Output aigogo-specific dependency groups (copy into your pyproject.toml)
aigg show-deps .aigogo/imports/aigogo/embedding_search --format poetry

# Then install with:
poetry install --with aigogo
```

**pip (requirements.txt):**
```bash
aigg show-deps .aigogo/imports/aigogo/embedding_search --format requirements | pip install -r /dev/stdin
```

**uv:**
```bash
# uv is a fast Python package manager (drop-in pip replacement)
aigg show-deps .aigogo/imports/aigogo/embedding_search --format requirements | uv pip install -r /dev/stdin

# Or add to an existing uv project
aigg show-deps .aigogo/imports/aigogo/embedding_search --format requirements | xargs uv add
```

### JavaScript Dependencies

```bash
# Install the snippet
aigg add token-budget-js:1.0.0
aigg install
```

**npm:**
```bash
# Output as package.json fragment with aigogo metadata (merge into your package.json)
aigg show-deps .aigogo/imports/@aigogo/token-budget-js --format npm

# The "aigogo" key in the output tracks which deps are aigogo-managed
# To remove later: delete listed packages and the "aigogo" key
```

**yarn:**
```bash
# Output as a ready-to-run yarn add command
aigg show-deps .aigogo/imports/@aigogo/token-budget-js --format yarn

# Or run it directly
eval "$(aigg show-deps .aigogo/imports/@aigogo/token-budget-js --format yarn)"
```

### show-deps Format Reference

| Format | Alias | Language | Output |
|--------|-------|----------|--------|
| `text` | | Any | Human-readable summary |
| `requirements` | `pip` | Python | `package>=1.0.0` (one per line, aigogo-labeled) |
| `pyproject` | `pep621` | Python | `[project.optional-dependencies] aigogo = [...]` TOML |
| `poetry` | | Python | `[tool.poetry.group.aigogo.dependencies]` TOML |
| `npm` | `package-json` | JavaScript | `{"dependencies": {...}, "aigogo": {...}}` JSON |
| `yarn` | | JavaScript | `yarn add "pkg@version"` commands |

This is a deliberate design choice -- aigg distributes reusable code, not full packages with dependency trees.

## AI Metadata

Every example includes an `ai` field in `aigogo.json`. This allows AI agents to discover and evaluate packages programmatically. See [MACHINES.md](../MACHINES.md) for the full specification.
