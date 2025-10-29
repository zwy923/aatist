from fastapi import APIRouter
router = APIRouter()

@router.get("/")
def list_projects():
    return [
        {"id": 1, "title": "Brand Logo", "tags": ["design", "branding"]},
        {"id": 2, "title": "Data Dashboard", "tags": ["react", "python"]},
    ]
