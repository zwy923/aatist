from pydantic import BaseModel
import os

class Settings(BaseModel):
    ENV: str = os.getenv("ENV", "development")
    DATABASE_URL: str = os.getenv("DATABASE_URL", "postgresql+psycopg2://aatist:aatist@localhost:5432/aatist")
    JWT_SECRET: str = os.getenv("JWT_SECRET", "dev_secret")
    JWT_ALGORITHM: str = os.getenv("JWT_ALGORITHM", "HS256")

settings = Settings()
