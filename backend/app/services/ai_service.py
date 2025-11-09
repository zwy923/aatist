from typing import Optional
import httpx
from app.core.config import settings


class AIService:
    """AI 业务逻辑层 - 负责与 recommender 微服务通信"""
    
    def __init__(self):
        self.recommender_base_url = getattr(settings, "RECOMMENDER_URL", "http://recommender:8001")
    
    def event_qa(self, query: str) -> dict:
        """
        轻量级 AI 问答 - 直接调用本地 AI 模块
        适合快速响应的 FAQ 类问题
        """
        from app.ai.event_qa_manager import get_event_qa_chain
        qa_chain = get_event_qa_chain()
        result = qa_chain.invoke(query)
        return {
            "question": query,
            "answer": result.content,
            "source": "local_ai"
        }
    
    async def get_recommendations(self, user_id: int, limit: int = 10) -> dict:
        """
        获取推荐 - 调用 recommender 微服务
        适合计算密集型的推荐任务
        """
        async with httpx.AsyncClient() as client:
            try:
                response = await client.post(
                    f"{self.recommender_base_url}/recommender/query",
                    json={"user_id": user_id, "limit": limit},
                    timeout=10.0
                )
                response.raise_for_status()
                return response.json()
            except httpx.RequestError as e:
                raise Exception(f"无法连接到推荐服务: {str(e)}")
    
    async def trigger_reindex(self) -> dict:
        """
        触发重新索引 - 调用 recommender 微服务
        用于定期更新向量索引
        """
        async with httpx.AsyncClient() as client:
            try:
                response = await client.post(
                    f"{self.recommender_base_url}/recommender/reindex",
                    timeout=30.0
                )
                response.raise_for_status()
                return response.json()
            except httpx.RequestError as e:
                raise Exception(f"无法连接到推荐服务: {str(e)}")


