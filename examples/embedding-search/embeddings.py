"""Minimal vector similarity search for embeddings."""

import math


def cosine_similarity(a, b):
    """Compute cosine similarity between two vectors.

    Args:
        a: List of floats.
        b: List of floats (same length as a).

    Returns:
        Float between -1.0 and 1.0.
    """
    if len(a) != len(b):
        raise ValueError(f"Dimension mismatch: {len(a)} vs {len(b)}")

    dot = sum(x * y for x, y in zip(a, b))
    norm_a = math.sqrt(sum(x * x for x in a))
    norm_b = math.sqrt(sum(x * x for x in b))

    if norm_a == 0 or norm_b == 0:
        return 0.0
    return dot / (norm_a * norm_b)


def top_k(query, corpus, k=5):
    """Find the k most similar items to a query vector.

    Args:
        query: Query embedding (list of floats).
        corpus: List of (id, embedding) tuples.
        k: Number of results to return.

    Returns:
        List of (id, score) tuples, sorted by descending similarity.
    """
    scored = [(item_id, cosine_similarity(query, emb)) for item_id, emb in corpus]
    scored.sort(key=lambda x: x[1], reverse=True)
    return scored[:k]


def batch_similarity(queries, corpus, k=5):
    """Run top_k for multiple queries.

    Args:
        queries: List of (query_id, embedding) tuples.
        corpus: List of (id, embedding) tuples.
        k: Number of results per query.

    Returns:
        Dict mapping query_id to list of (id, score) tuples.
    """
    return {qid: top_k(qemb, corpus, k) for qid, qemb in queries}


def deduplicate(items, threshold=0.95):
    """Remove near-duplicate items based on embedding similarity.

    Args:
        items: List of (id, embedding) tuples.
        threshold: Similarity threshold above which items are considered duplicates.

    Returns:
        List of (id, embedding) tuples with duplicates removed (keeps first seen).
    """
    unique = []
    for item_id, emb in items:
        is_dup = False
        for _, existing_emb in unique:
            if cosine_similarity(emb, existing_emb) >= threshold:
                is_dup = True
                break
        if not is_dup:
            unique.append((item_id, emb))
    return unique
