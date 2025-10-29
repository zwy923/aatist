from pydantic import BaseModel

class ProjectCreate(BaseModel):
    title: str
    description: str
    tags: str = ""

class ProjectOut(BaseModel):
    id: int
    title: str
    description: str
    tags: str

    class Config:
        from_attributes = True
