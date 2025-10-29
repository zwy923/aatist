from fastapi import APIRouter, Query

router = APIRouter()

@router.get("/match")
def match(project_id: int = Query(..., gt=0)):
    # stub response for MVP
    return {
        "project_id": project_id,
        "recommended_students": [
            {"id": 2, "name": "Bob", "score": 0.92},
            {"id": 5, "name": "Carol", "score": 0.88},
        ],
    }
