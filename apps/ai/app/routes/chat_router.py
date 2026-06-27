from fastapi import APIRouter, Depends
from pydantic import BaseModel
from app.core.chat.embedding import embedding
from app.core.chat.chat import ChatService, get_chat_service
from app.core.chat.rerank import rerank
from app.models.chat import ChatEmbeddingRequest, ChatRequest, ChatReqModel, ChatImageRequest, ChatAudioRequest


chat_router = APIRouter()


class RerankCandidate(BaseModel):
    source_id: str
    passage: str = ""


class RerankRequest(BaseModel):
    query: str
    candidates: list[RerankCandidate]


@chat_router.post("/chat/rerank")
def rerank_products(req: RerankRequest):
    """Cross-encoder rerank hybrid product candidates. Returns source_ids ordered
    most- to least-relevant. The Go worker keeps the top few and injects them."""
    items = [c.model_dump() for c in req.candidates]
    ranked = rerank(req.query, items, top_n=len(items))
    return {"source_ids": [c["source_id"] for c in ranked]}


@chat_router.post("/chat")
def chat(req: ChatRequest, service: ChatService = Depends(get_chat_service)):
    return {"chunks": service.handle(req.business_id, req.message)}


@chat_router.post("/chat/reply")
def chat_reply(req: ChatReqModel, service: ChatService = Depends(get_chat_service)):
    reply = service.handle_with_context(req)
    return {"reply": reply.message, "images": reply.images}


@chat_router.post("/chat/embed")
def embed_message(req: ChatEmbeddingRequest):
    chunks = embedding(req.message, "passage")
    return {"embeddings": [c.dict() for c in chunks]}

@chat_router.post("/chat/images")
def describe_image(req: ChatImageRequest, service: ChatService = Depends(get_chat_service)):
    description = service.handle_images(req)
    return {"description": description}


@chat_router.post("/chat/audio")
def transcribe_audio(req: ChatAudioRequest, service: ChatService = Depends(get_chat_service)):
    transcript = service.handle_audio(req)
    return {"transcript": transcript}
