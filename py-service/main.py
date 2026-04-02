import os
import logging
from fastapi import FastAPI
from fastapi.responses import JSONResponse
import httpx
from dotenv import load_dotenv
import json as json_lib

load_dotenv()

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

GO_SERVICE_URL = os.getenv("GO_SERVICE_URL", "http://localhost:8080")
HTTP_TIMEOUT = float(os.getenv("HTTP_TIMEOUT", "3.0"))

app = FastAPI(
    title="FastAPI Proxy for Go Service",
    description="Прокси-сервис для вызова Go микросервиса",
    version="1.0.0"
)

async def get_response_data(response):
    """Извлекает данные из ответа, обрабатывая как синхронные, так и асинхронные json методы"""
    if hasattr(response.json, '__call__'):
        # Если json - это метод, вызываем его
        result = response.json()
        # Если результат - корутина (async), ожидаем её
        if hasattr(result, '__await__'):
            result = await result
        return result
    return response.json

@app.get("/api/users/{user_id}")
async def get_user(user_id: str):
    """Получить пользователя из Go сервиса"""
    logger.info(f"GET /api/users/{user_id}")
    
    async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
        try:
            response = await client.get(f"{GO_SERVICE_URL}/users/{user_id}")
            data = await get_response_data(response)
            return JSONResponse(status_code=response.status_code, content=data)
        except httpx.TimeoutException:
            return JSONResponse(status_code=504, content={"error": f"Go service timeout after {HTTP_TIMEOUT}s"})
        except httpx.ConnectError:
            return JSONResponse(status_code=503, content={"error": f"Go service unavailable at {GO_SERVICE_URL}"})
        except Exception as e:
            logger.error(f"Unexpected error: {str(e)}")
            return JSONResponse(status_code=500, content={"error": f"Internal proxy error: {str(e)}"})

@app.post("/api/users")
async def create_user(user_data: dict):
    """Создать пользователя через Go сервис"""
    logger.info(f"POST /api/users")
    
    if "name" not in user_data:
        return JSONResponse(status_code=422, content={"error": "Missing required field: name"})
    
    async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
        try:
            response = await client.post(f"{GO_SERVICE_URL}/users", json=user_data)
            data = await get_response_data(response)
            return JSONResponse(status_code=response.status_code, content=data)
        except httpx.TimeoutException:
            return JSONResponse(status_code=504, content={"error": f"Go service timeout after {HTTP_TIMEOUT}s"})
        except httpx.ConnectError:
            return JSONResponse(status_code=503, content={"error": f"Go service unavailable at {GO_SERVICE_URL}"})
        except Exception as e:
            logger.error(f"Unexpected error: {str(e)}")
            return JSONResponse(status_code=500, content={"error": f"Internal proxy error: {str(e)}"})

@app.get("/api/health")
async def health():
    """Проверка здоровья обоих сервисов"""
    logger.info("GET /api/health")
    
    async with httpx.AsyncClient(timeout=2.0) as client:
        try:
            response = await client.get(f"{GO_SERVICE_URL}/health")
            go_healthy = response.status_code == 200
        except:
            go_healthy = False
    
    return {
        "fastapi": "healthy",
        "go_service": "healthy" if go_healthy else "unhealthy"
    }

@app.get("/")
async def root():
    return {
        "service": "FastAPI Proxy for Go Service",
        "endpoints": {
            "get_user": "GET /api/users/{user_id}",
            "create_user": "POST /api/users",
            "health": "GET /api/health",
            "docs": "GET /docs"
        }
    }