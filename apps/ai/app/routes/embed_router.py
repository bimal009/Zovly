from typing import Literal, Optional
from fastapi import APIRouter
from pydantic import BaseModel

from app.core.chat.embedding import embedding, EmbeddedChunk
from app.core.chat.helpers.text_formater import text_formatter
from app.models.faq import FaqRequest

embed_router = APIRouter()


@embed_router.post("/embed/faq", response_model=list[EmbeddedChunk])
def embed_faq(req: FaqRequest):
    text = text_formatter(req)
    return embedding(text, "passage")




class EmbedRequest(BaseModel):
    text: str
    kind: Literal["passage", "query"] = "passage"
    # chunk=False embeds the whole text as a single vector (atomic items like
    # products). prefix is prepended to chunks only if a long passage is split.
    chunk: bool = True
    prefix: Optional[str] = None

@embed_router.post("/embed", response_model=list[EmbeddedChunk])
def embed_text(req: EmbedRequest):
    return embedding(req.text, req.kind, req.chunk, req.prefix)