from typing import Optional, List
from sqlalchemy.orm import Session
from app.models.user import User
from app.repositories.base import BaseRepository


class UserRepository(BaseRepository[User]):
    """User 数据访问层"""
    
    def __init__(self, db: Session):
        super().__init__(User, db)
    
    def get_by_email(self, email: str) -> Optional[User]:
        """根据邮箱获取用户"""
        return self.db.query(User).filter(User.email == email).first()
    
    def get_by_skill(self, skill: str, skip: int = 0, limit: int = 100) -> List[User]:
        """根据技能获取用户"""
        return self.db.query(User).filter(User.skills.ilike(f"%{skill}%")).offset(skip).limit(limit).all()


