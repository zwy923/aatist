"""
API 路由聚合 - 统一管理所有 API 版本
"""
from fastapi import APIRouter
from app.api.v1 import events, projects, students, ai_router

api_router = APIRouter()

# 注册 v1 路由
api_router.include_router(events.router, prefix="/v1/events", tags=["Events"])
api_router.include_router(projects.router, prefix="/v1/projects", tags=["Projects"])
api_router.include_router(students.router, prefix="/v1/students", tags=["Students"])
api_router.include_router(ai_router.router, prefix="/v1/ai", tags=["AI"])


