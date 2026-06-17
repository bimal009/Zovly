from fastapi import APIRouter, Depends
from app.core.chat.chat import ChatService, get_chat_service
from app.models.chat import ChatRequest

chat_router = APIRouter()


@chat_router.post("/chat")
def chat(req: ChatRequest, service: ChatService = Depends(get_chat_service)):
    return {"chunks": service.handle(req.business_id, req.message)}
