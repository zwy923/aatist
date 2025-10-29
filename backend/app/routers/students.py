from fastapi import APIRouter
router = APIRouter()

# MVP static list
@router.get("/")
def list_students():
    return [
        {"id": 1, "name": "Alice", "skills": ["Design", "UI/UX"]},
        {"id": 2, "name": "Bob", "skills": ["Python", "ML"]},
    ]
