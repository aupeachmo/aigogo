"use strict";

/**
 * Estimate the number of tokens in a string.
 * Uses the ~4 characters-per-token heuristic typical of GPT-class models.
 * For precise counting, swap in a real tokenizer (e.g. gpt-tokenizer).
 * @param {string} text - The text to tokenize.
 * @returns {number} Estimated token count.
 */
function countTokens(text) {
  if (!text) return 0;
  return Math.ceil(text.length / 4);
}

/**
 * Check whether a message list fits within a token budget.
 * @param {Array<{role: string, content: string}>} messages - Chat messages.
 * @param {number} maxTokens - Maximum allowed tokens.
 * @returns {{fits: boolean, total: number, remaining: number}}
 */
function checkBudget(messages, maxTokens) {
  // Each message has ~4 tokens of overhead (role, delimiters)
  const MESSAGE_OVERHEAD = 4;
  let total = 0;
  for (const msg of messages) {
    total += countTokens(msg.content) + MESSAGE_OVERHEAD;
  }
  return {
    fits: total <= maxTokens,
    total,
    remaining: Math.max(0, maxTokens - total),
  };
}

/**
 * Split text into chunks that each fit within a token limit.
 * Splits on sentence boundaries when possible.
 * @param {string} text - The text to chunk.
 * @param {number} maxTokensPerChunk - Maximum tokens per chunk.
 * @returns {string[]} Array of text chunks.
 */
function chunkText(text, maxTokensPerChunk) {
  const sentences = text.match(/[^.!?]+[.!?]+\s*/g) || [text];
  const chunks = [];
  let current = "";

  for (const sentence of sentences) {
    const combined = current + sentence;
    if (countTokens(combined) > maxTokensPerChunk && current.length > 0) {
      chunks.push(current.trim());
      current = sentence;
    } else {
      current = combined;
    }
  }

  if (current.trim().length > 0) {
    chunks.push(current.trim());
  }

  return chunks;
}

/**
 * Trim a conversation to fit within a token budget by removing oldest messages.
 * The first message (system prompt) is always preserved.
 * @param {Array<{role: string, content: string}>} messages - Chat messages.
 * @param {number} maxTokens - Maximum allowed tokens.
 * @returns {Array<{role: string, content: string}>} Trimmed messages.
 */
function trimConversation(messages, maxTokens) {
  if (messages.length === 0) return [];

  const system = messages[0].role === "system" ? messages[0] : null;
  const rest = system ? messages.slice(1) : [...messages];

  let result = system ? [system] : [];
  let budget = checkBudget(result, maxTokens);

  // Add messages from newest to oldest
  const toAdd = [];
  for (let i = rest.length - 1; i >= 0; i--) {
    const candidate = [rest[i], ...toAdd];
    const test = [...result, ...candidate];
    const check = checkBudget(test, maxTokens);
    if (check.fits) {
      toAdd.unshift(rest[i]);
    } else {
      break;
    }
  }

  return [...result, ...toAdd];
}

module.exports = {
  countTokens,
  checkBudget,
  chunkText,
  trimConversation,
};
