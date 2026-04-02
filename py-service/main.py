import os
import logging
from fastapi import FastAPI, HTTPException, status
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
from typing import Optional, Dict, Any
import httpx
from dotenv import load_dotenv

load_dotenv()

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

GO_SERVICE_URL = os.getenv("GO_SERVICE_URL")
HTTP_TIMEOUT = float(os.getenv("HTTP_TIMEOUT", "3.0"))

# Pydantic модели для документации
class UserCreate(BaseModel):
    """Модель для создания пользователя"""
    name: str = Field(..., example="Alice", description="Имя пользователя")
    
    class Config:
        json_schema_extra = {
            "example": {
                "name": "Alice"
            }
        }

class UserResponse(BaseModel):
    """Модель ответа с пользователем"""
    id: str = Field(..., example="123", description="ID пользователя")
    name: str = Field(..., example="Alice", description="Имя пользователя")
    
    class Config:
        json_schema_extra = {
            "example": {
                "id": "123",
                "name": "Alice"
            }
        }

class ErrorResponse(BaseModel):
    """Модель ошибки"""
    error: str = Field(..., example="User not found", description="Описание ошибки")

class HealthResponse(BaseModel):
    """Модель проверки здоровья"""
    fastapi: str = Field(..., example="healthy", description="Статус FastAPI")
    go_service: str = Field(..., example="healthy", description="Статус Go сервиса")
    go_error: Optional[str] = Field(None, description="Ошибка Go сервиса если есть")
    go_url: str = Field(..., example="http://localhost:8080", description="URL Go сервиса")

app = FastAPI(
    title="FastAPI Proxy for Go Service",
    description="""
    # Прокси-сервис для вызова Go микросервиса
    
    Этот сервис предоставляет API для взаимодействия с Go микросервисом.
    
    ## Возможности:
    * Получение пользователя по ID
    * Создание нового пользователя
    * Проверка здоровья обоих сервисов
    
    ## Обработка ошибок:
    * `504` - Таймаут Go сервиса
    * `503` - Go сервис недоступен
    * `422` - Ошибка валидации входных данных
    """,
    version="1.0.0",
    contact={
        "name": "Lab-10 Team",
        "email": "team@example.com",
    },
    license_info={
        "name": "MIT",
    }
)

async def get_response_data(response):
    """Извлекает данные из ответа"""
    if hasattr(response.json, '__call__'):
        result = response.json()
        if hasattr(result, '__await__'):
            result = await result
        return result
    return response.json

@app.get(
    "/api/users/{user_id}",
    response_model=UserResponse,
    responses={
        200: {"description": "Пользователь найден", "model": UserResponse},
        404: {"description": "Пользователь не найден", "model": ErrorResponse},
        504: {"description": "Таймаут Go сервиса", "model": ErrorResponse},
        503: {"description": "Go сервис недоступен", "model": ErrorResponse},
    },
    summary="Получить пользователя",
    description="Возвращает пользователя по указанному ID из Go сервиса"
)
async def get_user(user_id: str):
    """
    Получить пользователя из Go сервиса.
    
    - **user_id**: Уникальный идентификатор пользователя (строка)
    
    ### Пример ответа:
    ```json
    {
        "id": "42",
        "name": "User 42"
    }
    ```
    """
    logger.info(f"GET /api/users/{user_id}")
    
    async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
        try:
            response = await client.get(f"{GO_SERVICE_URL}/users/{user_id}")
            data = await get_response_data(response)
            return JSONResponse(status_code=response.status_code, content=data)
        except httpx.TimeoutException:
            return JSONResponse(
                status_code=504, 
                content={"error": f"Go service timeout after {HTTP_TIMEOUT}s"}
            )
        except httpx.ConnectError:
            return JSONResponse(
                status_code=503, 
                content={"error": f"Go service unavailable at {GO_SERVICE_URL}"}
            )
        except Exception as e:
            logger.error(f"Unexpected error: {str(e)}")
            return JSONResponse(
                status_code=500, 
                content={"error": f"Internal proxy error: {str(e)}"}
            )

@app.post(
    "/api/users",
    response_model=UserResponse,
    status_code=status.HTTP_201_CREATED,
    responses={
        201: {"description": "Пользователь создан", "model": UserResponse},
        422: {"description": "Ошибка валидации", "model": ErrorResponse},
        504: {"description": "Таймаут Go сервиса", "model": ErrorResponse},
        503: {"description": "Go сервис недоступен", "model": ErrorResponse},
    },
    summary="Создать пользователя",
    description="Создаёт нового пользователя через Go сервис"
)
async def create_user(user_data: UserCreate):
    """
    Создать пользователя через Go сервис.
    
    - **user_data**: JSON объект с полем `name`
    
    ### Пример запроса:
    ```json
    {
        "name": "Bob"
    }
    ```
    
    ### Пример ответа:
    ```json
    {
        "id": "123",
        "name": "Bob"
    }
    ```
    """
    logger.info(f"POST /api/users - {user_data.name}")
    
    async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
        try:
            response = await client.post(
                f"{GO_SERVICE_URL}/users", 
                json=user_data.model_dump()
            )
            data = await get_response_data(response)
            return JSONResponse(status_code=response.status_code, content=data)
        except httpx.TimeoutException:
            return JSONResponse(
                status_code=504, 
                content={"error": f"Go service timeout after {HTTP_TIMEOUT}s"}
            )
        except httpx.ConnectError:
            return JSONResponse(
                status_code=503, 
                content={"error": f"Go service unavailable at {GO_SERVICE_URL}"}
            )
        except Exception as e:
            logger.error(f"Unexpected error: {str(e)}")
            return JSONResponse(
                status_code=500, 
                content={"error": f"Internal proxy error: {str(e)}"}
            )

@app.get(
    "/api/health",
    response_model=HealthResponse,
    responses={
        200: {"description": "Статус сервисов", "model": HealthResponse},
    },
    summary="Проверка здоровья",
    description="Проверяет статус FastAPI и Go сервисов"
)
async def health():
    """
    Проверка здоровья обоих сервисов.
    
    Возвращает статус каждого сервиса.
    """
    logger.info("GET /api/health")
    
    async with httpx.AsyncClient(timeout=2.0) as client:
        try:
            response = await client.get(f"{GO_SERVICE_URL}/health")
            go_healthy = response.status_code == 200
            go_error = None
        except Exception as e:
            go_healthy = False
            go_error = str(e)
    
    return {
        "fastapi": "healthy",
        "go_service": "healthy" if go_healthy else "unhealthy",
        "go_error": go_error,
        "go_url": GO_SERVICE_URL
    }

@app.get(
    "/",
    summary="Информация о сервисе",
    description="Возвращает информацию о доступных эндпоинтах"
)
async def root():
    return {
        "service": "FastAPI Proxy for Go Service",
        "version": "1.0.0",
        "endpoints": {
            "get_user": "GET /api/users/{user_id}",
            "create_user": "POST /api/users",
            "health": "GET /api/health",
            "docs": "GET /docs",
            "redoc": "GET /redoc"
    }
}