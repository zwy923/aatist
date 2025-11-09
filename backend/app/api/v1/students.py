from fastapi import APIRouter, Depends, Query, HTTPException
from typing import List, Optional
from app.services.user_service import UserService
from app.services.logger import get_logger
from app.schemas.user_schema import UserCreate, UserOut
from app.api.deps import get_user_service

router = APIRouter()


@router.get("/", response_model=List[UserOut])
def list_students(
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=100),
    skill: Optional[str] = None,
    user_service: UserService = Depends(get_user_service),
    logger = Depends(get_logger)
):
    """获取学生列表"""
    logger.log_info("backend", f"获取学生列表: skip={skip}, limit={limit}")
    
    if skill:
        users = user_service.get_users_by_skill(skill, skip=skip, limit=limit)
    else:
        users = user_service.get_users(skip=skip, limit=limit)
    
    return users


@router.get("/{student_id}", response_model=UserOut)
def get_student(
    student_id: int,
    user_service: UserService = Depends(get_user_service),
    logger = Depends(get_logger)
):
    """获取单个学生"""
    user = user_service.get_user(student_id)
    if not user:
        raise HTTPException(status_code=404, detail="学生不存在")
    
    logger.log_info("backend", f"获取学生: {student_id}")
    return user


@router.post("/", response_model=UserOut, status_code=201)
def create_student(
    user_data: UserCreate,
    user_service: UserService = Depends(get_user_service),
    logger = Depends(get_logger)
):
    """创建学生"""
    try:
        user = user_service.create_user(user_data)
        logger.log_action("backend", f"创建学生: {user.id}", {"email": user.email})
        return user
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.put("/{student_id}", response_model=UserOut)
def update_student(
    student_id: int,
    user_data: UserCreate,
    user_service: UserService = Depends(get_user_service),
    logger = Depends(get_logger)
):
    """更新学生"""
    try:
        user = user_service.update_user(student_id, user_data)
        if not user:
            raise HTTPException(status_code=404, detail="学生不存在")
        
        logger.log_action("backend", f"更新学生: {student_id}")
        return user
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.delete("/{student_id}", status_code=204)
def delete_student(
    student_id: int,
    user_service: UserService = Depends(get_user_service),
    logger = Depends(get_logger)
):
    """删除学生"""
    success = user_service.delete_user(student_id)
    if not success:
        raise HTTPException(status_code=404, detail="学生不存在")
    
    logger.log_action("backend", f"删除学生: {student_id}")
    return None


