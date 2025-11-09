from fastapi import APIRouter, Query
from backend.app.ai.event_qa_manager import get_event_qa_chain

router = APIRouter(prefix="/ai", tags=["AI"])

@router.get("/event_qa")
def event_qa(query: str = Query(...)):
    qa = get_event_qa_chain()
    result = qa.invoke(query)
    return {"question": query, "answer": result.content}

