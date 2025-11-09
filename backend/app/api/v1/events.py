from fastapi import APIRouter, Depends, Query, HTTPException
from typing import List, Optional
from app.services.event_service import EventService
from app.services.logger import get_logger
from app.schemas.event_schema import EventCreate, EventOut
from app.api.deps import get_event_service

router = APIRouter()


@router.get("/", response_model=List[EventOut])
def list_events(
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=100),
    category: Optional[str] = None,
    search: Optional[str] = None,
    event_service: EventService = Depends(get_event_service),
    logger = Depends(get_logger)
):
    """获取事件列表"""
    logger.log_info("backend", f"获取事件列表: skip={skip}, limit={limit}")
    
    if category:
        events = event_service.get_events_by_category(category, skip=skip, limit=limit)
    elif search:
        events = event_service.search_events(search, skip=skip, limit=limit)
    else:
        events = event_service.get_events(skip=skip, limit=limit)
    
    return events


@router.get("/{event_id}", response_model=EventOut)
def get_event(
    event_id: int,
    event_service: EventService = Depends(get_event_service),
    logger = Depends(get_logger)
):
    """获取单个事件"""
    event = event_service.get_event(event_id)
    if not event:
        raise HTTPException(status_code=404, detail="事件不存在")
    
    logger.log_info("backend", f"获取事件: {event_id}")
    return event


@router.post("/", response_model=EventOut, status_code=201)
def create_event(
    event_data: EventCreate,
    event_service: EventService = Depends(get_event_service),
    logger = Depends(get_logger)
):
    """创建事件"""
    event = event_service.create_event(event_data)
    logger.log_action("backend", f"创建事件: {event.id}", {"title": event.title})
    return event


@router.put("/{event_id}", response_model=EventOut)
def update_event(
    event_id: int,
    event_data: EventCreate,
    event_service: EventService = Depends(get_event_service),
    logger = Depends(get_logger)
):
    """更新事件"""
    event = event_service.update_event(event_id, event_data)
    if not event:
        raise HTTPException(status_code=404, detail="事件不存在")
    
    logger.log_action("backend", f"更新事件: {event_id}")
    return event


@router.delete("/{event_id}", status_code=204)
def delete_event(
    event_id: int,
    event_service: EventService = Depends(get_event_service),
    logger = Depends(get_logger)
):
    """删除事件"""
    success = event_service.delete_event(event_id)
    if not success:
        raise HTTPException(status_code=404, detail="事件不存在")
    
    logger.log_action("backend", f"删除事件: {event_id}")
    return None


