"""Sliding-window context manager for LLM conversations."""


def estimate_tokens(text):
    """Estimate token count using the ~4 chars per token heuristic.

    For precise counts, use tiktoken. This is a fast approximation.
    """
    return max(1, len(text) // 4)


class ContextWindow:
    """Manages a conversation history within a token budget.

    Keeps the system message pinned and trims older user/assistant
    turns when the budget is exceeded.
    """

    def __init__(self, max_tokens=4096, reserve_tokens=512, token_counter=None):
        """Initialize the context window.

        Args:
            max_tokens: Total token budget for the conversation.
            reserve_tokens: Tokens reserved for the next response.
            token_counter: Optional callable(str) -> int for precise token counting.
                           Defaults to the ~4 chars/token estimate.
        """
        self.max_tokens = max_tokens
        self.reserve_tokens = reserve_tokens
        self.count_tokens = token_counter or estimate_tokens
        self.system = None
        self.messages = []

    def set_system(self, content):
        """Set or replace the system message (always retained)."""
        self.system = {"role": "system", "content": content}

    def add(self, role, content):
        """Add a message to the conversation."""
        self.messages.append({"role": role, "content": content})

    def add_user(self, content):
        """Add a user message."""
        self.add("user", content)

    def add_assistant(self, content):
        """Add an assistant message."""
        self.add("assistant", content)

    def render(self):
        """Return the message list, trimmed to fit within the token budget.

        Trims the oldest non-system messages first. The system message
        and the most recent messages are always preserved.
        """
        budget = self.max_tokens - self.reserve_tokens
        result = []
        system_tokens = 0

        if self.system:
            system_tokens = self._message_tokens(self.system)
            budget -= system_tokens

        # Walk backwards from most recent, collecting messages that fit
        kept = []
        used = 0
        for msg in reversed(self.messages):
            msg_tokens = self._message_tokens(msg)
            if used + msg_tokens > budget:
                break
            kept.append(msg)
            used += msg_tokens

        kept.reverse()

        if self.system:
            result.append(self.system)
        result.extend(kept)
        return result

    def token_count(self):
        """Return the current total token count."""
        total = 0
        if self.system:
            total += self._message_tokens(self.system)
        for msg in self.messages:
            total += self._message_tokens(msg)
        return total

    def available_tokens(self):
        """Return how many tokens are available for the next response."""
        rendered = self.render()
        used = sum(self._message_tokens(m) for m in rendered)
        return self.max_tokens - used

    def clear(self, keep_system=True):
        """Clear conversation history."""
        self.messages = []
        if not keep_system:
            self.system = None

    def _message_tokens(self, msg):
        """Estimate tokens for a single message (content + overhead)."""
        return self.count_tokens(msg["content"]) + 4  # role + formatting overhead
