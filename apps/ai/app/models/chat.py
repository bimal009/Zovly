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
    context: ChatContext

ChatReqModel = ChatReplyRequest


class ChatIntentReqModel(BaseModel):
    message: str
    recent_turns: list[str] = Field(default_factory=list)






IntentType = Literal[
    "product_search",
    "service_inquiry",
    "event_inquiry",
    "order_status",
    "policy_faq",
    "booking_confirm",
    "cart_action",
    "small_talk",
    "complaint",
    "unknown",
]


class ExtractedEntities(BaseModel):
    item: Optional[str] = None
    color: Optional[str] = None
    size: Optional[str] = None
    budget_min: Optional[float] = None
    budget_max: Optional[float] = None
    quantity: Optional[int] = None
    date: Optional[str] = None
    order_id: Optional[str] = None
    product_ref: Optional[str] = None


class SubIntent(BaseModel):
    intent: IntentType
    text: str
    entities: ExtractedEntities = Field(default_factory=ExtractedEntities)
    confidence: float = Field(ge=0.0, le=1.0)


class IntentExtractionResult(BaseModel):
    sub_intents: list[SubIntent]

class AgentReply(BaseModel):
    message: str = Field(description="The reply text to send to the customer, plain text, a few sentences max.")
    images: list[str] = Field(
        default_factory=list,
        description="Image URLs to send with the reply. Only URLs that appear in the provided product context. Empty list if none.",
    )