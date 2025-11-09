from typing import List, Optional
from sqlalchemy.orm import Session
from app.repositories.project_repo import ProjectRepository
from app.schemas.project_schema import ProjectCreate, ProjectOut


class ProjectService:
    """Project 业务逻辑层"""
    
    def __init__(self, db: Session):
        self.repo = ProjectRepository(db)
    
    def get_project(self, project_id: int) -> Optional[ProjectOut]:
        """获取单个项目"""
        project = self.repo.get(project_id)
        return ProjectOut.model_validate(project) if project else None
    
    def get_projects(self, skip: int = 0, limit: int = 100) -> List[ProjectOut]:
        """获取项目列表"""
        projects = self.repo.get_all(skip=skip, limit=limit)
        return [ProjectOut.model_validate(project) for project in projects]
    
    def get_projects_by_tag(self, tag: str, skip: int = 0, limit: int = 100) -> List[ProjectOut]:
        """根据标签获取项目"""
        projects = self.repo.get_by_tags(tag, skip=skip, limit=limit)
        return [ProjectOut.model_validate(project) for project in projects]
    
    def search_projects(self, title: str, skip: int = 0, limit: int = 100) -> List[ProjectOut]:
        """搜索项目"""
        projects = self.repo.search_by_title(title, skip=skip, limit=limit)
        return [ProjectOut.model_validate(project) for project in projects]
    
    def create_project(self, project_data: ProjectCreate) -> ProjectOut:
        """创建项目"""
        project_dict = project_data.model_dump()
        project = self.repo.create(project_dict)
        return ProjectOut.model_validate(project)
    
    def update_project(self, project_id: int, project_data: ProjectCreate) -> Optional[ProjectOut]:
        """更新项目"""
        project = self.repo.get(project_id)
        if not project:
            return None
        project_dict = project_data.model_dump(exclude_unset=True)
        updated_project = self.repo.update(project, project_dict)
        return ProjectOut.model_validate(updated_project)
    
    def delete_project(self, project_id: int) -> bool:
        """删除项目"""
        return self.repo.delete(project_id)


