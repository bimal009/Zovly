from typing import Literal
from pydantic import BaseModel
from app.core.chat.chunking import chunk_document
from langchain_huggingface import HuggingFaceEmbeddings

model = HuggingFaceEmbeddings(
    model_name="intfloat/multilingual-e5-large",
    encode_kwargs={"normalize_embeddings": True},
)

class EmbeddedChunk(BaseModel):
    chunk_index: int
    content: str
    embedding: list[float]

def embedding(
    text: str,
    kind: Literal["passage", "query"] = "passage",
) -> list[EmbeddedChunk]:
    chunks = [text] if kind == "query" else chunk_document(text)
    vectors = model.embed_documents([f"{kind}: {c}" for c in chunks])
    return [
        EmbeddedChunk(chunk_index=i, content=c, embedding=v)
        for i, (c, v) in enumerate(zip(chunks, vectors))
    ]