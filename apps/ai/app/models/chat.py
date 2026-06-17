from typing import Literal
from pydantic import BaseModel

class ChatRequest(BaseModel):
    business_id: str
    platform: Literal['instagram', 'facebook', 'whatsapp', 'tiktok']
    thread_id: str | None = None 
    message: str