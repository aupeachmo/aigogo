---
theme: default
title: "aigogo: Lightweight Agent Distribution"
info: |
  ## aigogo
  A lightweight agent manager.

  ML/AI Melbourne — February 2026
class: text-center
drawings:
  persist: false
transition: slide-left
mdc: true
---

# aigogo

## Make packaging and distributing your AI agents a breeze

<div class="pt-12">
  <span class="px-2 py-1 rounded cursor-pointer" hover="bg-white bg-opacity-10">
    ML/AI Melbourne — February 2026
  </span>
</div>

<div class="abs-br m-6 flex gap-2">
  <a href="https://github.com/aupeachmo/aigogo" target="_blank" alt="GitHub" title="Open in GitHub"
    class="text-xl slidev-icon-btn opacity-50 !border-none !hover:text-white">
    <carbon-logo-github />
  </a>
</div>

<!--
Welcome everyone. Tonight I want to show you a tool I've been building called aigogo. It solves a problem that I think everyone in this room has experienced.
-->

---

# whoami

<div class="grid grid-cols-2 gap-12 mt-8 items-center">

<div>

<div class="text-4xl font-bold">Dushan Karovich-Wynne</div>
<div class="text-xl text-gray-400 mt-2">Founder — subrosa.ai</div>

<div class="mt-10 space-y-4 text-lg">
  <div class="flex items-center gap-3">
    <carbon-logo-github class="text-2xl" />
    <a href="https://github.com/aupeachmo">github.com/aupeachmo</a>
  </div>
  <div class="flex items-center gap-3">
    <carbon-logo-linkedin class="text-2xl" />
    <a href="https://linkedin.com/in/dushankw">linkedin.com/in/dushankw</a>
  </div>
  <div class="flex items-center gap-3">
    <carbon-logo-x class="text-2xl" />
    <span>@aupeachmo</span>
  </div>
</div>

</div>

<div class="text-center">

<img src="/qr.png" class="mx-auto rounded" style="width: 200px" />
<div class="text-gray-400 mt-4 text-sm">github.com/aupeachmo/aigogo</div>

</div>

</div>

---
transition: fade-out
---

# The Copy-Paste Problem

You build something useful for your AI project:

- A prompt template with variable substitution
- A response parser that handles messy LLM output
- An entire agent, eg: interface with an external CRM

It works great. Then you need it in **another project**.

```
project-a/
  └── utils/prompt_templates.py  ← original

project-b/
  └── utils/prompt_templates.py  ← copy #1

project-c/
  └── utils/prompt_templates.py  ← copy #2 (outdated)

project-d/
  └── utils/prompt_templates.py  ← copy #3 (has a bug fix the others don't)
```

<div class="text-center text-xl mt-4">
  5 people × 8 projects × 3 utility modules = <span class="text-red-400 font-bold">chaos</span>
</div>

<!--
Raise your hand if you've ever copy-pasted a utility module between projects.

You write a prompt template or a tool-calling decorator and it works great. Then your colleague needs it, or you start a new project. So you copy it over. Now you have two copies. You fix a bug in one — but which copy did you update?

Multiply this across a team and it becomes unmanageable very quickly.
-->

---

# Why Not Just Use...?

<div class="grid grid-cols-2 gap-8 mt-8">

<div class="p-4 rounded bg-red-900 bg-opacity-20 border border-red-800">

### PyPI / npm

Publishing overhead: accounts, CI pipelines, versioning ceremony

**Too heavy** for many use cases

</div>

<div class="p-4 rounded bg-red-900 bg-opacity-20 border border-red-800">

### Git Submodules

Fragile merges, confusing commands, version pinning is painful

Everyone has been **burned by submodules**

</div>

<div class="p-4 rounded bg-red-900 bg-opacity-20 border border-red-800">

### Monorepo

Not always feasible across teams and orgs, breaks down across identity boundaries

**Tight coupling** between unrelated projects

</div>

<div class="p-4 rounded bg-red-900 bg-opacity-20 border border-red-800">

### Private Package Registry

Cost, setup, maintenance overhead

**Overkill** + running one for each language ecosystem you use

</div>

</div>

<div class="text-center mt-8 text-lg">
These tools distribute <span class="text-yellow-400 font-bold">libraries</span>. We need something for <span class="text-green-400 font-bold">agents</span>.
</div>

<!--
So why can't we use existing tools?

PyPI is great but it's designed for libraries with proper release processes. Setting up a publishing pipeline for a single utility file is massive overkill.

Git submodules — I think everyone in this room has a war story about submodules. Fragile merges, confusing commands, painful versioning.

Monorepos work for some orgs but they force tight coupling and aren't always feasible across teams.

Private package registries cost money and need maintenance. Overkill for "I just want to share a decorator."

The fundamental issue: these tools were designed for distributing libraries. What we need is something designed for small, self-contained pieces of AI code.
-->

---

# Introducing aigogo

<div class="text-2xl text-center mt-4 mb-8 text-green-400">
  "Package it once, share it everywhere"
</div>

- **Agent manager**, not a package manager
- Single binary, written in Go, zero runtime dependencies
- Uses **Docker registries** as transport
- Works with Python, JavaScript/TypeScript (...for now)
- Lock files for reproducibility
- Content-addressable storage for integrity

<div class="mt-6 p-4 bg-blue-900 bg-opacity-30 rounded border border-blue-700">

**Install:**
```bash
brew install aupeachmo/tap/aigogo
```

</div>

<!--
aigogo is a lightweight agent manager — not a package manager, that distinction matters.

It's a single Go binary with zero dependencies. You install it and you're done.

The key insight, which I'll explain in a moment, is that it uses Docker registries as its transport layer. You don't need to set up any infrastructure. Docker registries are everywhere and they're free for public packages.
-->

---

# Docker Registries as Transport

<div class="mt-8">

Docker registries are **everywhere**, **free**, and **battle-tested**:

*You can easily run private ones*

<div class="grid grid-cols-3 gap-4 mt-6 text-center">
<div class="p-3 rounded bg-gray-800">Docker Hub</div>
<div class="p-3 rounded bg-gray-800">ghcr.io</div>
<div class="p-3 rounded bg-gray-800">AWS ECR</div>
<div class="p-3 rounded bg-gray-800">Google GCR</div>
<div class="p-3 rounded bg-gray-800">Azure ACR</div>
<div class="p-3 rounded bg-gray-800">GitLab Registry</div>
</div>

<div class="mt-8 p-4 bg-green-900 bg-opacity-30 rounded border border-green-700 text-lg text-center">
  <strong>No Docker daemon needed.</strong><br/>
  aigogo speaks the Docker V2 Registry HTTP API directly.
</div>

</div>

<!--
Here's the clever part. Docker registries already solve storage, versioning, authentication, and namespacing. They're available everywhere — Docker Hub, GitHub, AWS, Google Cloud, Azure, GitLab, self-hosted — and they're free for public packages.

And importantly: aigogo does NOT require Docker to be installed. It speaks the Docker V2 Registry HTTP API directly. There's no Docker daemon, no container runtime. It's just HTTP calls to upload and download your agent code.
-->

---

# How It Works

<div class="mt-4">

### Author Workflow

```bash
aigg init                    # Create aigogo.json manifest
aigg build                   # Package to local cache
aigg push registry/pkg:1.0   # Upload to registry
```

### Consumer Workflow

```bash
aigg add registry/pkg:1.0    # Pull + store by SHA256 hash
aigg install                 # Create import symlinks
```

### Then Just Import

```python
from aigogo.my_utils import helper_function
```

```javascript
const utils = require('@aigogo/my-utils')
```

</div>

<!--
The workflow is five commands total. Three for the author: init, build, push. Two for the consumer: add, install.

After install, you just import. Python uses a namespace — from aigogo dot package name. JavaScript uses scoped packages — at aigogo slash package name. No PATH hacking, no sys.path manipulation. It just works.
-->

---

# Under the Hood

<div class="grid grid-cols-2 gap-6 mt-6">

<div>

### Content-Addressable Storage

```
~/.aigogo/store/sha256/
  └── ab/abc123.../
      ├── files/         # Read-only
      └── aigogo.json    # Manifest
```

- Immutable storage by **SHA256 hash**
- Files made read-only after storage
- Integrity verification on every install

</div>

<div>

### Lock Files

```json
{
  "packages": {
    "tool_use_decorator": {
      "version": "1.0.0",
      "integrity": "sha256:abc12...",
      "source": "docker.io/org/pkg:1.0.0",
      "language": "python"
    }
  }
}
```

- Pin exact versions + hashes
- **Commit to git** — teammates run `aigg install`
- Familiar model (npm, pip, cargo)

</div>

</div>

<div class="mt-4 p-3 bg-gray-800 rounded text-center">
  <strong>No dependency resolution by design.</strong> aigogo manages source code.<br/>
  Your existing package manager (pip, npm) handles dependency trees.
</div>

<!--
Under the hood, packages are stored in a content-addressable store, keyed by SHA256 hash. This means storage is immutable and integrity is always verified.

Lock files pin exact versions and integrity hashes — you commit the lock file to git, and your teammates just run aigg install. This is the same model as npm's package-lock or pip's requirements.txt, so it's immediately familiar.

One deliberate design decision: aigogo does NOT do dependency resolution. It manages source code. If your package needs numpy, you declare that in the manifest and aigogo can generate the right requirements.txt, but pip still handles installing numpy. This keeps aigogo simple and avoids reinventing the wheel.
-->

---

# Namespace Imports — No PATH Hacks

<div class="grid grid-cols-2 gap-6 mt-4">

<div>

### Python

```
.aigogo/imports/aigogo/
├── __init__.py
└── tool_use_decorator/ → symlink
```

Uses Python's `.pth` mechanism — the same thing pip uses internally.

Works with **system Python, venv, Poetry, and uv**.

</div>

<div>

### JavaScript

```
.aigogo/imports/@aigogo/
└── my-utils/
    ├── index.js        → symlink
    └── package.json    # auto-generated
```

Uses `NODE_PATH` via a `register.js` loader. Works with CommonJS and ESM.

</div>

</div>

<div class="mt-4 p-3 bg-gray-800 rounded text-center">
  The <code>.aigogo/</code> directory is gitignored. Commit <code>aigogo.lock</code> — regenerate with <code>aigg install</code>.
</div>

<!--
The import mechanism is designed to feel native. In Python, aigogo creates symlinks under a namespace package and uses Python's .pth file mechanism, which is the same thing pip uses internally. So "from aigogo dot tool use decorator import tool" just works, in any Python environment — venv, Poetry, uv, or system Python.

For JavaScript, it uses scoped packages under @aigogo and sets up NODE_PATH. Standard require or import syntax works.

The .aigogo directory is gitignored. You commit the lock file and regenerate the imports directory with aigg install. Clean and reproducible.
-->

---
layout: center
class: text-center
---

# Demo Time

<div class="text-2xl mt-4 text-gray-400">
  Author → Registry → Consumer → Import
</div>

<div class="text-lg mt-8 text-gray-500">
  Full lifecycle in one workflow
</div>

<!--
OK, let's see this in action. I'm going to walk through the complete lifecycle: taking some code, packaging it, pushing it to a registry, then pulling it into a fresh project and using it.
-->

---

# Demo: The Package

<div class="mt-2">

A tool-calling decorator — turns Python functions into OpenAI tool schemas:

```python {all|1-2|4-5|7-13|15-17}{maxHeight:'380px'}
from aigogo.tool_use_decorator import tool, get_tools, call_tool

# Decorate any function — schema is auto-generated from type hints
@tool
def search_docs(query: str, limit: int = 5) -> list:
    """Search documentation for relevant sections.

    Args:
        query: The search query string.
        limit: Maximum number of results to return.
    """
    return [f"Result for: {query}"]

# Pass to OpenAI API
tools = get_tools()
# → [{"type": "function", "function": {"name": "search_docs", ...}}]
```

</div>

<v-click>

<div class="mt-4 p-3 bg-yellow-900 bg-opacity-30 rounded border border-yellow-700 text-center">
  Zero dependencies. Single file. Useful across every LLM project.
  <br/>
  <strong>This is exactly the kind of code you end up copy-pasting.</strong>
</div>

</v-click>

<!--
Here's the package we'll work with: a tool-calling decorator. You decorate a Python function and it auto-generates an OpenAI-compatible tool schema from the type hints and docstring.

This is exactly the kind of utility you build once, find incredibly useful, and then end up copying into every project. Zero dependencies, single file, universally needed.

[SWITCH TO TERMINAL FOR LIVE DEMO]
-->

---

# Demo: Author Workflow

````md magic-move
```bash
# Look at the manifest
cat aigogo.json
```

```json
{
  "name": "tool-use-decorator",
  "version": "1.0.0",
  "description": "Convert Python functions into LLM tool-calling schemas",
  "language": { "name": "python", "version": ">=3.8,<4.0" },
  "files": { "include": ["tools.py"] },
  "ai": {
    "summary": "Auto-generate OpenAI-compatible tool schemas from type hints",
    "capabilities": [
      "Convert decorated functions into tool-calling JSON schemas",
      "Maintain a registry of all decorated tools",
      "Dispatch tool calls by name with argument parsing"
    ]
  }
}
```

```bash
# Build locally — no Docker needed
aigg build
# → Built tool-use-decorator:1.0.0 to ~/.aigogo/cache/

# Push to a registry
aigg push docker.io/myorg/tool-use-decorator:1.0.0 \
  --from tool-use-decorator:1.0.0
# → Pushed successfully
```
````

<!--
[LIVE DEMO - narrate each step]

First, the manifest. It describes the package — name, version, language, which files to include. Notice the "ai" field at the bottom — I'll come back to that.

Then we build. This packages the files into the local cache. No Docker daemon needed, this is all local.

Then we push to Docker Hub. The --from flag tells aigogo which local build to upload.
-->

---

# Demo: Consumer Workflow

````md magic-move
```bash
# In a fresh project...
mkdir demo-project && cd demo-project

# Add the package — pulls from registry, stores by hash
aigg add docker.io/myorg/tool-use-decorator:1.0.0
```

```bash
# Lock file created automatically
cat aigogo.lock
{
  "packages": {
    "tool_use_decorator": {
      "version": "1.0.0",
      "integrity": "sha256:a1b2c3d4e5...",
      "source": "docker.io/myorg/tool-use-decorator:1.0.0"
    }
  }
}
```

```bash
# Install — creates import symlinks
aigg install
# → Installed tool_use_decorator

# Now just use it
python3 -c "
from aigogo.tool_use_decorator import tool, get_tools
import json

@tool
def summarize(text: str, max_words: int = 100) -> str:
    '''Summarize text to a target length.'''
    return text[:max_words]

print(json.dumps(get_tools(), indent=2))
"
```
````

<!--
[LIVE DEMO continued]

Now I switch to a completely fresh directory. aigg add pulls the package from the registry and stores it by its SHA256 hash.

It automatically creates a lock file — you commit this to git. Your teammates just run aigg install.

aigg install creates the symlink structure. And now — the payoff — I can import it with a standard Python import. No sys.path hacking, no PYTHONPATH, nothing. Just "from aigogo dot tool_use_decorator import tool".

And it works. The schema is auto-generated from the type hints. This is the complete lifecycle — from code to registry to importable module.
-->

---

# AI Agent Discovery (beta)

The `ai` field in `aigogo.json` — metadata for autonomous agents:

<div class="grid grid-cols-2 gap-6 mt-4">

```json
{
  "ai": {
    "summary": "Auto-generate OpenAI tool
      schemas from type hints",
    "capabilities": [
      "Convert functions into schemas",
      "Maintain a tool registry",
      "Dispatch tool calls by name"
    ],
    "usage": "from aigogo.tool_use_decorator
      import tool\n\n@tool\ndef my_fn()...",
    "inputs": "Python functions with type hints",
    "outputs": "OpenAI-compatible tool schemas"
  }
}
```

<div class="mt-2">

### How Agents Use This

1. **Search** registries for packages by capability
2. **Read** `summary` to decide relevance
3. **Copy** `usage` to write integration code
4. **No human needed** in the loop

<div class="mt-4 p-3 bg-purple-900 bg-opacity-30 rounded border border-purple-700">
  Agents discovering and composing other agents — <strong>programmatically</strong>.
</div>

</div>

</div>

<!--
Here's where it gets interesting for this audience. The ai field in the manifest is optional metadata designed for autonomous agents, not humans.

An agent can search a registry, read the summary to decide if a package is relevant, look at the capabilities list, and copy the usage example to write integration code — all without a human in the loop.

This is the vision: agents that can discover, evaluate, and compose other agents programmatically. We're not fully there yet — there's no cross-registry search index — but the metadata layer is ready.
-->

---

# The Claude Code Skill

aigogo ships with a Claude Code slash command:

```
.claude/commands/aigogo.md
```

Type `/aigogo` in Claude Code and the AI walks you through:

<div class="grid grid-cols-3 gap-4 mt-4">
<div class="p-3 rounded bg-gray-800 text-center">
  <div class="text-green-400 text-xl mb-2">Create</div>
  init → detect language → add files → validate → build
</div>
<div class="p-3 rounded bg-gray-800 text-center">
  <div class="text-blue-400 text-xl mb-2">Consume</div>
  add → install → show import syntax
</div>
<div class="p-3 rounded bg-gray-800 text-center">
  <div class="text-purple-400 text-xl mb-2">Publish</div>
  validate → build → test → login → push
</div>
</div>

<div class="mt-6 p-3 bg-gray-800 rounded">

Copy it to your own project:
```bash
cp .claude/commands/aigogo.md your-project/.claude/commands/
```

Now your AI assistant knows how to use aigogo in your project.

</div>

<!--
aigogo also ships with a Claude Code slash command. You type /aigogo and the AI assistant walks you through the entire workflow interactively — creating packages, consuming them, or publishing.

You can copy this command file into any of your projects. Now your AI assistant knows how to use aigogo without you having to explain anything.

This is two layers of AI integration: metadata for autonomous agents, and a skill for AI-assisted human workflows.
-->

---

# Team Workflow

<div class="grid grid-cols-2 gap-12 mt-8">

<div class="border-r border-gray-700 pr-8">

### Author

Package and publish your agent code.

```bash
aigg build
aigg push org/pkg:1.0
git commit aigogo.lock
git push
```

Lock file captures version + SHA256 integrity hash.

</div>

<div>

### Teammate

Pull and start using immediately.

```bash
git pull
aigg install
```

```python
from aigogo.pkg import ...
# Just works
```

One command to sync. Fully deterministic.

</div>

</div>

<!--
Here's what the team workflow looks like. The author builds, pushes to a registry, and commits the lock file to git.

A teammate pulls the repo, runs aigg install, and the imports are ready. One command to sync, completely deterministic because the lock file pins exact versions and integrity hashes.
-->

---

# Design Decisions

<div class="mt-4">

| Decision | Rationale |
|----------|-----------|
| **No Docker daemon** | HTTP API only. No heavy dependency. |
| **No dependency resolution** | Source code, not libraries. Let pip/npm do their job. |
| **Content-addressable storage** | Immutability + integrity by default. |
| **Symlinks for imports** | Zero-copy. Native paths. Instant install. |
| **Single Go binary** | No runtime. Cross-platform. 2 external deps total. |

</div>

<!--
A few design decisions worth highlighting.

No Docker daemon — we just speak HTTP to the registry. No dependency resolution — we deliberately avoid that complexity and let your existing package manager handle it. Content-addressable storage for integrity. Symlinks for zero-copy, instant installs.

The entire codebase has only two external Go dependencies. This is a tool that practices what it preaches about keeping things lightweight.
-->

---

# Getting Started

<div class="grid grid-cols-2 gap-8 mt-8">

<div>

### Install

```bash
# macOS / Linux
brew install aupeachmo/tap/aigogo

# Or download binary
# github.com/aupeachmo/aigogo/releases

# Or from source
go install github.com/aupeachmo/aigogo@latest
```

### Try the Examples

```bash
git clone github.com/aupeachmo/aigogo
cd aigogo/examples/tool-use-decorator
aigg build
```

</div>

<div>

### 6 Example Packages

- **prompt-templates** — Variable substitution + chaining
- **llm-response-parser** — Extract JSON from messy LLM output
- **tool-use-decorator** — Auto-generate tool schemas
- **agent-context-manager** — Sliding-window conversation memory
- **embedding-search** — Vector similarity search
- **token-budget-js** — Token counting (JavaScript)

</div>

</div>

<!--
Getting started is straightforward. Brew install on Mac or Linux, or download a binary, or go install from source.

The repo comes with six example packages that cover common AI utility patterns — prompt templates, response parsing, tool schemas, conversation memory, embedding search, and token budgeting. Each one is a real, useful package you can build and push right now.
-->

---
layout: center
class: text-center
---

# Thank You

<div class="text-2xl mt-4">
  <carbon-logo-github class="inline" /> <a href="https://github.com/aupeachmo/aigogo">github.com/aupeachmo/aigogo</a>
</div>

<div class="mt-8 text-gray-400">

```bash
brew install aupeachmo/tap/aigogo
```

</div>

<div class="mt-12 text-lg text-gray-500">
  Questions?
</div>

<div class="mt-6 text-sm text-gray-600">
  Feedback and contributions welcome — issues and PRs on GitHub.
</div>

<!--
Thank you. The repo is on GitHub, and I'd love to hear what you think. Questions?
-->
