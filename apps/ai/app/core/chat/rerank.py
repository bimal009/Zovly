import logging

from sentence_transformers import CrossEncoder

logger = logging.getLogger(__name__)


_reranker = CrossEncoder("BAAI/bge-reranker-v2-m3", max_length=512)


def rerank(query: str, candidates: list[dict], top_n: int = 8) -> list[dict]:
    """Reorder hybrid-search candidates by cross-encoder relevance to the query.

    Each candidate is a dict with at least `passage` (falls back to `name`).
    Returns the top `top_n` candidates in descending relevance order. On any
    failure the original order is preserved (truncated to top_n) so search never
    hard-fails on the reranker.
    """
    if not candidates:
        return []

    try:
        pairs = [(query, c.get("passage") or c.get("name") or "") for c in candidates]
        scores = _reranker.predict(pairs)
        ranked = sorted(zip(candidates, scores), key=lambda cs: cs[1], reverse=True)
        return [c for c, _ in ranked[:top_n]]
    except Exception:
        logger.exception("rerank failed; falling back to hybrid order")
        return candidates[:top_n]
