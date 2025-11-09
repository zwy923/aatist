from fastapi import APIRouter, Depends, Query, HTTPException
from app.services.ai_service import AIService
from app.services.logger import get_logger
from app.api.deps import get_ai_service

router = APIRouter()


@router.get("/event_qa")
def event_qa(
    query: str = Query(..., description="问题内容"),
    ai_service: AIService = Depends(get_ai_service),
    logger = Depends(get_logger)
):
    """
    事件问答 - 轻量级 AI 功能
    使用本地 AI 模块进行快速问答
    """
    try:
        logger.log_info("ai", f"事件问答请求: {query}")
        result = ai_service.event_qa(query)
        logger.log_info("ai", "事件问答完成")
        return result
    except Exception as e:
        logger.log_error("ai", f"事件问答失败: {str(e)}")
        raise HTTPException(status_code=500, detail=f"AI 服务错误: {str(e)}")


@router.post("/recommendations")
async def get_recommendations(
    user_id: int = Query(..., description="用户ID"),
    limit: int = Query(10, ge=1, le=50, description="返回数量"),
    ai_service: AIService = Depends(get_ai_service),
    logger = Depends(get_logger)
):
    """
    获取推荐 - 调用 recommender 微服务
    计算密集型的推荐任务
    """
    try:
        logger.log_info("ai", f"获取推荐: user_id={user_id}, limit={limit}")
        result = await ai_service.get_recommendations(user_id, limit)
        logger.log_info("ai", "推荐获取完成")
        return result
    except Exception as e:
        logger.log_error("ai", f"获取推荐失败: {str(e)}")
        raise HTTPException(status_code=500, detail=f"推荐服务错误: {str(e)}")


@router.post("/reindex")
async def trigger_reindex(
    ai_service: AIService = Depends(get_ai_service),
    logger = Depends(get_logger)
):
    """
    触发重新索引 - 调用 recommender 微服务
    用于定期更新向量索引
    """
    try:
        logger.log_info("ai", "触发重新索引")
        result = await ai_service.trigger_reindex()
        logger.log_info("ai", "重新索引完成")
        return result
    except Exception as e:
        logger.log_error("ai", f"重新索引失败: {str(e)}")
        raise HTTPException(status_code=500, detail=f"索引服务错误: {str(e)}")


