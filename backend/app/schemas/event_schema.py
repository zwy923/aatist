from pydantic import BaseModel

class EventCreate(BaseModel):
    title: str
    category: str = "general"
    description: str = ""

class EventOut(BaseModel):
    id: int
    title: str
    category: str
    description: str

    class Config:
        from_attributes = True
