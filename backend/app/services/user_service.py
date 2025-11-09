from typing import List, Optional
from sqlalchemy.orm import Session
from app.repositories.user_repo import UserRepository
from app.schemas.user_schema import UserCreate, UserOut


class UserService:
    """User 业务逻辑层"""
    
    def __init__(self, db: Session):
        self.repo = UserRepository(db)
    
    def get_user(self, user_id: int) -> Optional[UserOut]:
        """获取单个用户"""
        user = self.repo.get(user_id)
        return UserOut.model_validate(user) if user else None
    
    def get_user_by_email(self, email: str) -> Optional[UserOut]:
        """根据邮箱获取用户"""
        user = self.repo.get_by_email(email)
        return UserOut.model_validate(user) if user else None
    
    def get_users(self, skip: int = 0, limit: int = 100) -> List[UserOut]:
        """获取用户列表"""
        users = self.repo.get_all(skip=skip, limit=limit)
        return [UserOut.model_validate(user) for user in users]
    
    def get_users_by_skill(self, skill: str, skip: int = 0, limit: int = 100) -> List[UserOut]:
        """根据技能获取用户"""
        users = self.repo.get_by_skill(skill, skip=skip, limit=limit)
        return [UserOut.model_validate(user) for user in users]
    
    def create_user(self, user_data: UserCreate) -> UserOut:
        """创建用户"""
        # 检查邮箱是否已存在
        existing_user = self.repo.get_by_email(user_data.email)
        if existing_user:
            raise ValueError(f"邮箱 {user_data.email} 已被使用")
        
        user_dict = user_data.model_dump()
        user = self.repo.create(user_dict)
        return UserOut.model_validate(user)
    
    def update_user(self, user_id: int, user_data: UserCreate) -> Optional[UserOut]:
        """更新用户"""
        user = self.repo.get(user_id)
        if not user:
            return None
        
        # 如果更新邮箱，检查是否冲突
        if user_data.email != user.email:
            existing_user = self.repo.get_by_email(user_data.email)
            if existing_user:
                raise ValueError(f"邮箱 {user_data.email} 已被使用")
        
        user_dict = user_data.model_dump(exclude_unset=True)
        updated_user = self.repo.update(user, user_dict)
        return UserOut.model_validate(updated_user)
    
    def delete_user(self, user_id: int) -> bool:
        """删除用户"""
        return self.repo.delete(user_id)


