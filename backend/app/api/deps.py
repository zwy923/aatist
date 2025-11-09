"""
依赖注入 - 提供数据库会话等服务依赖
"""
from fastapi import Depends
from sqlalchemy.orm import Session
from app.core.database import get_db
from app.services.event_service import EventService
from app.services.project_service import ProjectService
from app.services.user_service import UserService
from app.services.ai_service import AIService
from app.services.logger import logger


def get_event_service(db: Session = Depends(get_db)) -> EventService:
    """获取 Event 服务实例"""
    return EventService(db)


def get_project_service(db: Session = Depends(get_db)) -> ProjectService:
    """获取 Project 服务实例"""
    return ProjectService(db)


def get_user_service(db: Session = Depends(get_db)) -> UserService:
    """获取 User 服务实例"""
    return UserService(db)


def get_ai_service() -> AIService:
    """获取 AI 服务实例"""
    return AIService()


def get_logger():
    """获取日志服务实例"""
    return logger


