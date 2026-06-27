from typing import Literal, Optional
from pydantic import BaseModel,Field


class ChatRequest(BaseModel):
    business_id: str
    platform: Literal['instagram', 'facebook', 'whatsapp', 'tiktok']
    thread_id: Optional[str] = None
    message: str


class ChatEmbeddingRequest(BaseModel):
    message: str

class ChatImageRequest(BaseModel):
    url: str


class ChatAudioRequest(BaseModel):
    url: str


class KnowledgeChunk(BaseModel):
    content: str
    source_type: Optional[str] = None
    source_id: Optional[str] = None
    score: float = 0.0


class PastChatChunk(BaseModel):
    content: str
    conversation_id: Optional[str] = None
    score: float = 0.0


class BusinessInfo(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    website: Optional[str] = None
    phone: Optional[str] = None
    address: Optional[str] = None


class CustomerInfo(BaseModel):
    contact_name: Optional[str] = None
    contact_username: Optional[str] = None


class PastMessage(BaseModel):
    direction: str  # "in" | "out"
    content: Optional[str] = None


class ChatContext(BaseModel):
    knowledge: list[KnowledgeChunk] = []
    past_chats: list[PastChatChunk] = []
    business: Optional[BusinessInfo] = None
    past_conversation: list[PastMessage] = []
    customer: Optional[CustomerInfo] = None


class ChatReplyRequest(BaseModel):
    business_id: str
    conversation_id: Optional[str] = None  # used by tools to record the active product
    message: str
    active_product_id: Optional[str] = None  # product currently under discussion, if any
    context: ChatContext

ChatReqModel = ChatReplyRequest


class AgentReply(BaseModel):
    message: str = Field(description="The reply text to send to the customer, plain text, a few sentences max.")
    images: list[str] = Field(
        default_factory=list,
        description="Image URLs to send with the reply. Only URLs that appear in the provided product context. Empty list if none.",
    )