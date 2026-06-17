from app.config.db import get_db
from sqlalchemy import text
from fastapi import Depends
from app.core.chat.embedding import embedding


class ChatService:
    def __init__(self, db):
        self.db = db

    def handle(self, business_id: str, message: str) -> str:
        query_chunks = embedding(message, kind="query")
        query_vector = query_chunks[0].embedding

        knowledge = self.search_knowledge(business_id, query_vector, top_k=5)
        knowledge = [c for c in knowledge if c["score"] >= 0.80]

        past = self.search_past_messages(business_id, query_vector, top_k=3)
        past = [m for m in past if m["score"] >= 0.80]

        return {"knowledge": knowledge, "past_messages": past}

    def search_knowledge(self, business_id: str, query_vector: list[float], top_k: int = 5):
        sql = text("""
            SELECT content, source_type, 1 - (embedding <=> :vec) AS score
            FROM knowledge_chunks
            WHERE business_id = :business_id
            ORDER BY embedding <=> :vec
            LIMIT :top_k
        """)
        vec = "[" + ",".join(map(str, query_vector)) + "]"
        rows = self.db.execute(
            sql, {"vec": vec, "business_id": business_id, "top_k": top_k},
        ).fetchall()
        return [
            {"content": r[0], "source_type": r[1], "score": r[2]}
            for r in rows
        ]

    def search_past_messages(self, business_id: str, query_vector: list[float], top_k: int = 3):
        sql = text("""
            SELECT content, conversation_id, 1 - (embedding <=> :vec) AS score
            FROM message_embeddings
            WHERE business_id = :business_id
            ORDER BY embedding <=> :vec
            LIMIT :top_k
        """)
        vec = "[" + ",".join(map(str, query_vector)) + "]"
        rows = self.db.execute(
            sql, {"vec": vec, "business_id": business_id, "top_k": top_k},
        ).fetchall()
        return [
            {"content": r[0], "conversation_id": str(r[1]), "score": r[2]}
            for r in rows
        ]


def get_chat_service(db=Depends(get_db)) -> ChatService:
    return ChatService(db)