from sqlalchemy import Column, Integer, String, Text
from app.core.database import Base

class Event(Base):
    __tablename__ = "events"
    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(150), nullable=False)
    category = Column(String(50), default="general")
    description = Column(Text, default="")
