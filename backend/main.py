from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.api.router import api_router
from app.core.config import settings

app = FastAPI(title="aatist Backend API", version="0.1.0")

origins = [
    "http://localhost:5173",  # 前端地址
    "http://127.0.0.1:5173"
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,          # 允许的前端源
    allow_credentials=True,
    allow_methods=["*"],            # 允许所有 HTTP 方法
    allow_headers=["*"],            # 允许所有请求头
)

@app.get("/")
def root():
    return {"message": "aatist backend is running 🚀", "env": settings.ENV}

# 包含 API 路由
app.include_router(api_router, prefix="/api")
