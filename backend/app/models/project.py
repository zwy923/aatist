from sqlalchemy import Column, Integer, String, Text
from app.core.database import Base

class Project(Base):
    __tablename__ = "projects"
    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(150), nullable=False)
    description = Column(Text, nullable=False)
    tags = Column(String(255), default="")
