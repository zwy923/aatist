from fastapi import FastAPI, Body

app = FastAPI(title="aatist Recommender", version="0.1.0")

@app.get("/")
def root():
    return {"message": "recommender service is running 🧠"}

@app.post("/recommend")
def recommend(project: dict = Body(...)):
    # Very naive stub: echo skills and fake scores
    skills = project.get("skills", [])
    return {
        "project": project.get("title", "Untitled"),
        "skills": skills,
        "candidates": [
            {"id": 1, "name": "Alice", "score": 0.75},
            {"id": 2, "name": "Bob", "score": 0.72},
        ],
    }
