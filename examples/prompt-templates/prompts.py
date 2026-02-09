"""Structured prompt templates with variable substitution and chaining."""

from dataclasses import dataclass, field


@dataclass
class Prompt:
    """A prompt template with named placeholders."""

    system: str = ""
    user: str = ""
    variables: dict = field(default_factory=dict)

    def render(self, **kwargs):
        """Render the prompt, substituting variables.

        Variables passed as kwargs override defaults.
        """
        merged = {**self.variables, **kwargs}
        messages = []
        if self.system:
            messages.append({"role": "system", "content": self.system.format(**merged)})
        if self.user:
            messages.append({"role": "user", "content": self.user.format(**merged)})
        return messages

    def chain(self, other, **kwargs):
        """Chain this prompt's output as input to another prompt.

        Returns a function that takes the first response and renders the second.
        """
        def run(response_text):
            return other.render(previous=response_text, **kwargs)
        return run


# --- Built-in templates ---

CODE_REVIEW = Prompt(
    system="You are a senior {language} developer performing a code review.",
    user="Review this code for bugs, security issues, and style:\n\n```{language}\n{code}\n```",
    variables={"language": "python"},
)

SUMMARIZE = Prompt(
    system="You summarize text concisely. Target length: {length}.",
    user="Summarize:\n\n{text}",
    variables={"length": "2-3 sentences"},
)

EXTRACT_STRUCTURED = Prompt(
    system=(
        "Extract structured data from the input. "
        "Return valid JSON matching this schema:\n{schema}"
    ),
    user="{text}",
)

CHAIN_OF_THOUGHT = Prompt(
    system=(
        "Think step by step. First analyze the problem, "
        "then provide your reasoning, then give a final answer."
    ),
    user="{question}",
)

REFINE = Prompt(
    system="You are refining a previous response based on feedback.",
    user="Previous response:\n{previous}\n\nFeedback: {feedback}\n\nProvide an improved version.",
)
