from fastapi import APIRouter, Depends
from app.core.chat.embedding import embedding
from app.core.chat.chat import ChatService, get_chat_service
from app.models.chat import ChatEmbeddingRequest, ChatRequest, ChatReqModel

chat_router = APIRouter()


@chat_router.post("/chat")
def chat(req: ChatRequest, service: ChatService = Depends(get_chat_service)):
    return {"chunks": service.handle(req.business_id, req.message)}


@chat_router.post("/chat/reply")
def chat_reply(req: ChatReqModel, service: ChatService = Depends(get_chat_service)):
    reply = service.handle_with_context(req)
    return {"reply": reply}


@chat_router.post("/chat/embed")
def embed_message(req: ChatEmbeddingRequest):
    chunks = embedding(req.message, "passage")
    return {"embeddings": [c.dict() for c in chunks]}