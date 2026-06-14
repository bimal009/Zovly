from pydantic import BaseModel
from typing import Optional
from enum import Enum

class KnowledgeSourceType(str, Enum):
    faq = "faq"
    policy = "policy"
    post = "post"

class KnowledgeChunk(BaseModel):
    id: str
    business_id: str
    source_type: KnowledgeSourceType
    source_id: str
    chunk_index: int
    content: str
    embedding: list[float]
    metadata: Optional[dict] = None



class EmbeddedChunk(BaseModel):
    chunk_index: int
    content: str
    embedding: list[float]