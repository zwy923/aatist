"""
事件日志服务 - 使用 Redis Stream 异步推送日志
"""
import json
import redis
from typing import Optional
from app.core.config import settings


class LoggerService:
    """日志服务 - 通过 Redis Stream 异步推送日志到 event_logger"""
    
    def __init__(self):
        redis_url = getattr(settings, "REDIS_URL", "redis://localhost:6379/0")
        self.redis_client = redis.from_url(redis_url, decode_responses=True)
        self.stream_name = "event_logs"
    
    def push_event(self, service: str, event_type: str, message: str, metadata: Optional[dict] = None) -> bool:
        """
        推送事件到 Redis Stream
        
        Args:
            service: 来源服务 (e.g., "backend", "ai", "scraper")
            event_type: 事件类型 (e.g., "INFO", "ERROR", "ACTION", "USER_LOGIN")
            message: 事件消息
            metadata: 额外的元数据
        
        Returns:
            bool: 是否成功推送
        """
        try:
            event_data = {
                "service": service,
                "type": event_type,
                "message": message,
            }
            if metadata:
                event_data["metadata"] = json.dumps(metadata)
            
            # 推送到 Redis Stream
            message_id = self.redis_client.xadd(
                self.stream_name,
                event_data,
                maxlen=10000  # 限制流的最大长度
            )
            return True
        except Exception as e:
            # 如果 Redis 不可用，记录错误但不阻塞主流程
            print(f"⚠️ 日志推送失败: {str(e)}")
            return False
    
    def log_info(self, service: str, message: str, metadata: Optional[dict] = None):
        """记录信息日志"""
        return self.push_event(service, "INFO", message, metadata)
    
    def log_error(self, service: str, message: str, metadata: Optional[dict] = None):
        """记录错误日志"""
        return self.push_event(service, "ERROR", message, metadata)
    
    def log_action(self, service: str, action: str, metadata: Optional[dict] = None):
        """记录操作日志"""
        return self.push_event(service, "ACTION", action, metadata)


# 全局日志服务实例
logger = LoggerService()


