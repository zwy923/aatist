from typing import Optional, List
from sqlalchemy.orm import Session
from app.models.project import Project
from app.repositories.base import BaseRepository


class ProjectRepository(BaseRepository[Project]):
    """Project 数据访问层"""
    
    def __init__(self, db: Session):
        super().__init__(Project, db)
    
    def get_by_tags(self, tag: str, skip: int = 0, limit: int = 100) -> List[Project]:
        """根据标签获取项目"""
        return self.db.query(Project).filter(Project.tags.ilike(f"%{tag}%")).offset(skip).limit(limit).all()
    
    def search_by_title(self, title: str, skip: int = 0, limit: int = 100) -> List[Project]:
        """根据标题搜索项目"""
        return self.db.query(Project).filter(Project.title.ilike(f"%{title}%")).offset(skip).limit(limit).all()


