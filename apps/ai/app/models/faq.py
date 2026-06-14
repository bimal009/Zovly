from pydantic import BaseModel

class FaqRequest(BaseModel):
    question: str
    answer: str


