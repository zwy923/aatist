from fastapi import APIRouter
router = APIRouter()

@router.get("/")
def list_events():
    return [
        {"id": 1, "title": "Hackathon", "category": "innovation"},
        {"id": 2, "title": "Career Talk", "category": "career"},
    ]
