"""Convert Python functions into LLM tool-calling schemas."""

import inspect
import json
import re

# Python type hints to JSON Schema types
_TYPE_MAP = {
    "str": "string",
    "int": "integer",
    "float": "number",
    "bool": "boolean",
    "list": "array",
    "dict": "object",
    "None": "null",
}

_REGISTRY = {}


def tool(fn):
    """Decorator that registers a function as an LLM-callable tool.

    Extracts the function name, docstring, and type hints to build
    an OpenAI-compatible function-calling schema.

    Usage:
        @tool
        def get_weather(city: str, units: str = "celsius") -> str:
            '''Get the current weather for a city.

            Args:
                city: The city name to look up.
                units: Temperature units (celsius or fahrenheit).
            '''
            ...
    """
    schema = function_schema(fn)
    _REGISTRY[fn.__name__] = {"function": fn, "schema": schema}
    fn._tool_schema = schema
    return fn


def function_schema(fn):
    """Build an OpenAI function-calling schema from a Python function.

    Args:
        fn: A function with type hints and a docstring.

    Returns:
        Dict matching the OpenAI tool schema format.
    """
    sig = inspect.signature(fn)
    hints = fn.__annotations__
    doc = inspect.getdoc(fn) or ""

    # Parse description (first paragraph) and arg descriptions from docstring
    description, arg_docs = _parse_docstring(doc)

    properties = {}
    required = []

    for name, param in sig.parameters.items():
        prop = {}

        # Type from annotation
        if name in hints:
            type_name = getattr(hints[name], "__name__", str(hints[name]))
            prop["type"] = _TYPE_MAP.get(type_name, "string")

        # Description from docstring
        if name in arg_docs:
            prop["description"] = arg_docs[name]

        properties[name] = prop

        if param.default is inspect.Parameter.empty:
            required.append(name)

    return {
        "type": "function",
        "function": {
            "name": fn.__name__,
            "description": description,
            "parameters": {
                "type": "object",
                "properties": properties,
                "required": required,
            },
        },
    }


def get_tools():
    """Return all registered tool schemas as a list (for the API tools parameter)."""
    return [entry["schema"] for entry in _REGISTRY.values()]


def call_tool(name, arguments):
    """Call a registered tool by name with the given arguments.

    Args:
        name: The function name.
        arguments: Dict of keyword arguments, or a JSON string.

    Returns:
        The function's return value.
    """
    if name not in _REGISTRY:
        raise ValueError(f"Unknown tool: {name}")

    if isinstance(arguments, str):
        arguments = json.loads(arguments)

    return _REGISTRY[name]["function"](**arguments)


def _parse_docstring(doc):
    """Parse a Google-style docstring into description and arg docs."""
    lines = doc.strip().split("\n")

    description_lines = []
    arg_docs = {}
    in_args = False

    for line in lines:
        stripped = line.strip()
        if stripped.lower().startswith("args:"):
            in_args = True
            continue
        if stripped.lower().startswith(("returns:", "raises:", "yields:")):
            in_args = False
            continue
        if in_args:
            match = re.match(r"(\w+)\s*(?:\(.*?\))?\s*:\s*(.*)", stripped)
            if match:
                arg_docs[match.group(1)] = match.group(2).strip()
        elif not in_args and stripped:
            description_lines.append(stripped)

    description = " ".join(description_lines)
    return description, arg_docs
