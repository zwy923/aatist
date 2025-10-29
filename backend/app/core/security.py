from datetime import datetime, timedelta
from typing import Optional
import jwt
from app.core.config import settings

def create_token(sub: str, expires_delta: Optional[timedelta] = None) -> str:
    now = datetime.utcnow()
    expire = now + (expires_delta or timedelta(hours=12))
    payload = {"sub": sub, "iat": now, "exp": expire}
    token = jwt.encode(payload, settings.JWT_SECRET, algorithm=settings.JWT_ALGORITHM)
    return token
