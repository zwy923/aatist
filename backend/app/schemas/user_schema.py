from pydantic import BaseModel, EmailStr

class UserCreate(BaseModel):
    name: str
    email: EmailStr
    skills: str = ""

class UserOut(BaseModel):
    id: int
    name: str
    email: EmailStr
    skills: str

    class Config:
        from_attributes = True
