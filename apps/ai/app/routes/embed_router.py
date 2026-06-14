from typing import Literal
from fastapi import APIRouter

from app.core.chat.embedding import embedding, EmbeddedChunk
from app.core.chat.helpers.text_formater import text_formatter
from app.models.faq import FaqRequest

embed_router = APIRouter()


@embed_router.post("/embed/faq", response_model=list[EmbeddedChunk])
def embed_faq(req: FaqRequest):
    text = text_formatter(req)
    return embedding(text, "passage")
