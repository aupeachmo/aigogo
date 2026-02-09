"""Parse and validate structured output from LLM responses."""

import json
import re


def extract_json(text):
    """Extract JSON from an LLM response, handling markdown code blocks.

    Tries in order:
    1. JSON inside ```json ... ``` code blocks
    2. JSON inside ``` ... ``` code blocks
    3. First { ... } or [ ... ] in the raw text
    """
    # Try ```json blocks first
    match = re.search(r"```json\s*\n?(.*?)\n?\s*```", text, re.DOTALL)
    if match:
        return _parse_lenient(match.group(1).strip())

    # Try plain ``` blocks
    match = re.search(r"```\s*\n?(.*?)\n?\s*```", text, re.DOTALL)
    if match:
        candidate = match.group(1).strip()
        result = _try_parse(candidate)
        if result is not None:
            return result

    # Try raw JSON
    match = re.search(r"(\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}|\[.*?\])", text, re.DOTALL)
    if match:
        return _parse_lenient(match.group(1).strip())

    raise ValueError("No JSON found in response")


def extract_list(text, separator="\n"):
    """Extract a list of items from an LLM response.

    Handles numbered lists (1. item), bullet lists (- item, * item),
    and plain newline-separated items.
    """
    lines = text.strip().split(separator)
    items = []
    for line in lines:
        line = line.strip()
        # Strip numbering: "1. ", "2) ", etc.
        line = re.sub(r"^\d+[.)]\s*", "", line)
        # Strip bullets: "- ", "* ", "• "
        line = re.sub(r"^[-*•]\s*", "", line)
        line = line.strip()
        if line:
            items.append(line)
    return items


def validate_keys(data, required=None, optional=None):
    """Validate that parsed JSON has expected structure.

    Args:
        data: Parsed JSON (dict).
        required: Keys that must be present.
        optional: Keys that may be present. If set, extra keys raise ValueError.

    Returns:
        The validated data dict.
    """
    if not isinstance(data, dict):
        raise ValueError(f"Expected dict, got {type(data).__name__}")

    required = set(required or [])
    missing = required - set(data.keys())
    if missing:
        raise ValueError(f"Missing required keys: {missing}")

    if optional is not None:
        allowed = required | set(optional)
        extra = set(data.keys()) - allowed
        if extra:
            raise ValueError(f"Unexpected keys: {extra}")

    return data


def _parse_lenient(text):
    """Parse JSON leniently, fixing common LLM mistakes."""
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        pass

    # Fix trailing commas before } or ]
    fixed = re.sub(r",\s*([}\]])", r"\1", text)
    try:
        return json.loads(fixed)
    except json.JSONDecodeError:
        pass

    # Fix single quotes
    fixed = text.replace("'", '"')
    try:
        return json.loads(fixed)
    except json.JSONDecodeError:
        pass

    raise ValueError(f"Could not parse JSON: {text[:200]}")


def _try_parse(text):
    """Try to parse JSON, return None on failure."""
    try:
        return _parse_lenient(text)
    except ValueError:
        return None
