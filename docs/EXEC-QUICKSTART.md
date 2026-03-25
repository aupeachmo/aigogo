# exec — Quickstart

Run an agent directly with `aigg exec`. Unix/macOS only.

## 1. Create Your Agent

```python
# run.py
import sys
print(f"Hello from my-agent! Args: {sys.argv[1:]}")
```

## 2. Package It

```bash
aigg init
aigg add file run.py
```

Add a `scripts` field to `aigogo.json`:

```json
{
  "scripts": {
    "my-agent": "run.py"
  }
}
```

Build:

```bash
aigg build my-agent:1.0.0
```

## 3. Install It

```bash
aigg add my-agent:1.0.0
aigg install
```

## 4. Run It

```bash
aigg exec my-agent
aigg exec my-agent arg1 arg2
ENV_VAR=foo aigg exec my-agent
```

Dependencies are installed automatically on first run into an isolated environment (`~/.aigogo/envs/`).

## Platform Support

`aigg exec` requires Unix/macOS. It is not supported on Windows.
