from app.routes.embed_router import embed_router
from app.routes.chat_router import chat_router
from fastapi import APIRouter, FastAPI, Response

app = FastAPI(root_path="/api/v1")

ml_router = APIRouter(prefix="/ml")
ml_router.include_router(embed_router)
ml_router.include_router(chat_router)
# ml_router.include_router(search_router)
app.include_router(ml_router)

@app.get("/health")
def health():
    return Response(content="ok", status_code=200)