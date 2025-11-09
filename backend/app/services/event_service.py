from typing import List, Optional
from sqlalchemy.orm import Session
from app.repositories.event_repo import EventRepository
from app.schemas.event_schema import EventCreate, EventOut


class EventService:
    """Event 业务逻辑层"""
    
    def __init__(self, db: Session):
        self.repo = EventRepository(db)
    
    def get_event(self, event_id: int) -> Optional[EventOut]:
        """获取单个事件"""
        event = self.repo.get(event_id)
        return EventOut.model_validate(event) if event else None
    
    def get_events(self, skip: int = 0, limit: int = 100) -> List[EventOut]:
        """获取事件列表"""
        events = self.repo.get_all(skip=skip, limit=limit)
        return [EventOut.model_validate(event) for event in events]
    
    def get_events_by_category(self, category: str, skip: int = 0, limit: int = 100) -> List[EventOut]:
        """根据分类获取事件"""
        events = self.repo.get_by_category(category, skip=skip, limit=limit)
        return [EventOut.model_validate(event) for event in events]
    
    def search_events(self, title: str, skip: int = 0, limit: int = 100) -> List[EventOut]:
        """搜索事件"""
        events = self.repo.search_by_title(title, skip=skip, limit=limit)
        return [EventOut.model_validate(event) for event in events]
    
    def create_event(self, event_data: EventCreate) -> EventOut:
        """创建事件"""
        event_dict = event_data.model_dump()
        event = self.repo.create(event_dict)
        return EventOut.model_validate(event)
    
    def update_event(self, event_id: int, event_data: EventCreate) -> Optional[EventOut]:
        """更新事件"""
        event = self.repo.get(event_id)
        if not event:
            return None
        event_dict = event_data.model_dump(exclude_unset=True)
        updated_event = self.repo.update(event, event_dict)
        return EventOut.model_validate(updated_event)
    
    def delete_event(self, event_id: int) -> bool:
        """删除事件"""
        return self.repo.delete(event_id)


