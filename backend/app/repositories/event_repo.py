from typing import Optional, List
from sqlalchemy.orm import Session
from app.models.event import Event
from app.repositories.base import BaseRepository


class EventRepository(BaseRepository[Event]):
    """Event 数据访问层"""
    
    def __init__(self, db: Session):
        super().__init__(Event, db)
    
    def get_by_category(self, category: str, skip: int = 0, limit: int = 100) -> List[Event]:
        """根据分类获取事件"""
        return self.db.query(Event).filter(Event.category == category).offset(skip).limit(limit).all()
    
    def search_by_title(self, title: str, skip: int = 0, limit: int = 100) -> List[Event]:
        """根据标题搜索事件"""
        return self.db.query(Event).filter(Event.title.ilike(f"%{title}%")).offset(skip).limit(limit).all()


