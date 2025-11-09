from fastapi import APIRouter, Depends, Query, HTTPException
from typing import List, Optional
from app.services.project_service import ProjectService
from app.services.logger import get_logger
from app.schemas.project_schema import ProjectCreate, ProjectOut
from app.api.deps import get_project_service

router = APIRouter()


@router.get("/", response_model=List[ProjectOut])
def list_projects(
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=100),
    tag: Optional[str] = None,
    search: Optional[str] = None,
    project_service: ProjectService = Depends(get_project_service),
    logger = Depends(get_logger)
):
    """获取项目列表"""
    logger.log_info("backend", f"获取项目列表: skip={skip}, limit={limit}")
    
    if tag:
        projects = project_service.get_projects_by_tag(tag, skip=skip, limit=limit)
    elif search:
        projects = project_service.search_projects(search, skip=skip, limit=limit)
    else:
        projects = project_service.get_projects(skip=skip, limit=limit)
    
    return projects


@router.get("/{project_id}", response_model=ProjectOut)
def get_project(
    project_id: int,
    project_service: ProjectService = Depends(get_project_service),
    logger = Depends(get_logger)
):
    """获取单个项目"""
    project = project_service.get_project(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")
    
    logger.log_info("backend", f"获取项目: {project_id}")
    return project


@router.post("/", response_model=ProjectOut, status_code=201)
def create_project(
    project_data: ProjectCreate,
    project_service: ProjectService = Depends(get_project_service),
    logger = Depends(get_logger)
):
    """创建项目"""
    project = project_service.create_project(project_data)
    logger.log_action("backend", f"创建项目: {project.id}", {"title": project.title})
    return project


@router.put("/{project_id}", response_model=ProjectOut)
def update_project(
    project_id: int,
    project_data: ProjectCreate,
    project_service: ProjectService = Depends(get_project_service),
    logger = Depends(get_logger)
):
    """更新项目"""
    project = project_service.update_project(project_id, project_data)
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")
    
    logger.log_action("backend", f"更新项目: {project_id}")
    return project


@router.delete("/{project_id}", status_code=204)
def delete_project(
    project_id: int,
    project_service: ProjectService = Depends(get_project_service),
    logger = Depends(get_logger)
):
    """删除项目"""
    success = project_service.delete_project(project_id)
    if not success:
        raise HTTPException(status_code=404, detail="项目不存在")
    
    logger.log_action("backend", f"删除项目: {project_id}")
    return None


