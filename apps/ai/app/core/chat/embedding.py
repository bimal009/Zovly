from typing import Literal, Optional
from pydantic import BaseModel
from app.core.chat.chunking import chunk_document
from langchain_huggingface import HuggingFaceEmbeddings

model = HuggingFaceEmbeddings(
    model_name="intfloat/multilingual-e5-large",
    encode_kwargs={"normalize_embeddings": True},
)

# e5-large handles ~512 tokens. Passages shorter than this embed cleanly as a
# single vector; only genuinely long docs need chunking.
MAX_PASSAGE_CHARS = 1800


class EmbeddedChunk(BaseModel):
    chunk_index: int
    content: str
    embedding: list[float]


def embedding(
    text: str,
    kind: Literal["passage", "query"] = "passage",
    chunk: bool = True,
    prefix: Optional[str] = None,
) -> list[EmbeddedChunk]:
    """Embed text with the e5 query/passage prefix.

    chunk=False embeds the whole text as ONE vector (used for atomic items like
    products). If such a passage is unusually long it is split, but the product
    name (`prefix`) is prepended to every chunk so no chunk is orphaned.
    Queries are always single-vector regardless of `chunk`.
    """
    if kind == "query":
        pieces = [text]
    elif not chunk:
        if len(text) <= MAX_PASSAGE_CHARS:
            pieces = [text]
        else:
            pieces = chunk_document(text)
            if prefix:
                pieces = [f"{prefix}: {c}" for c in pieces]
    else:
        pieces = chunk_document(text)

    vectors = model.embed_documents([f"{kind}: {c}" for c in pieces])
    return [
        EmbeddedChunk(chunk_index=i, content=c, embedding=v)
        for i, (c, v) in enumerate(zip(pieces, vectors))
    ]
