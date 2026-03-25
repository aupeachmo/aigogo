# aigg exec — Quickstart

Make your agent executable in 4 steps.

## 1. Create your agent

```python
# run.py
import sys

def main():
    print("Hello from my agent")
    print("Args:", sys.argv[1:])

if __name__ == "__main__":
    main()
```

## 2. Package it

```bash
aigg init
aigg add file run.py
```

Add `scripts` to your `aigogo.json`:

```json
{
  "name": "my-agent",
  "version": "1.0.0",
  "language": { "name": "python", "version": ">=3.8" },
  "files": { "include": ["run.py"] },
  "scripts": { "my-agent": "run.py" }
}
```

Build:

```bash
aigg build
```

## 3. Install it (in any project)

```bash
aigg add my-agent:1.0.1
aigg install
```

## 4. Run it

```bash
aigg exec my-agent
aigg exec my-agent arg1 arg2
API_KEY=sk-123 aigg exec my-agent
```

That's it. Dependencies are installed automatically on first run.

## Cleanup

```bash
aigg clean          # see what's using disk
aigg clean --envs   # remove cached environments
```
