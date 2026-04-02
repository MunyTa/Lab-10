import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch, AsyncMock
import httpx
from main import app

client = TestClient(app)

# Тест 1: Успешный GET запрос к Go сервису
def test_proxy_get_user_success():
    """Тест успешного получения пользователя из Go сервиса"""
    mock_response = {
        "id": "123",
        "name": "User 123"
    }
    
    with patch('httpx.AsyncClient.get') as mock_get:
        mock_get.return_value = AsyncMock(
            status_code=200,
            json=AsyncMock(return_value=mock_response)
        )
        
        response = client.get("/api/users/123")
        
        assert response.status_code == 200
        assert response.json() == mock_response
        mock_get.assert_called_once()

# Тест 2: Go сервис возвращает 404
def test_proxy_get_user_404():
    """Тест обработки 404 от Go сервиса"""
    with patch('httpx.AsyncClient.get') as mock_get:
        mock_get.return_value = AsyncMock(
            status_code=404,
            json=AsyncMock(return_value={"error": "user not found"})
        )
        
        response = client.get("/api/users/999")
        
        assert response.status_code == 404
        assert "error" in response.json()

# Тест 3: Успешный POST запрос к Go сервису
def test_proxy_create_user_success():
    """Тест успешного создания пользователя через Go сервис"""
    user_data = {"name": "Alice"}
    mock_response = {"id": "123", "name": "Alice"}
    
    with patch('httpx.AsyncClient.post') as mock_post:
        mock_post.return_value = AsyncMock(
            status_code=201,
            json=AsyncMock(return_value=mock_response)
        )
        
        response = client.post("/api/users", json=user_data)
        
        assert response.status_code == 201
        assert response.json() == mock_response
        mock_post.assert_called_once()

# Тест 4: Таймаут при вызове Go сервиса
def test_go_service_timeout():
    """Тест обработки таймаута от Go сервиса"""
    with patch('httpx.AsyncClient.get') as mock_get:
        mock_get.side_effect = httpx.TimeoutException("Timeout")
        
        response = client.get("/api/users/123")
        
        assert response.status_code == 504
        assert "timeout" in response.json()["error"].lower()

# Тест 5: Go сервис недоступен
def test_go_service_unavailable():
    """Тест обработки недоступности Go сервиса"""
    with patch('httpx.AsyncClient.get') as mock_get:
        mock_get.side_effect = httpx.ConnectError("Connection refused")
        
        response = client.get("/api/users/123")
        
        assert response.status_code == 503
        assert "unavailable" in response.json()["error"].lower()

# Тест 6: Go сервис возвращает 500 ошибку
def test_go_service_internal_error():
    """Тест обработки 500 ошибки от Go сервиса"""
    with patch('httpx.AsyncClient.get') as mock_get:
        mock_get.return_value = AsyncMock(
            status_code=500,
            json=AsyncMock(return_value={"error": "internal server error"})
        )
        
        response = client.get("/api/users/123")
        
        assert response.status_code == 500
        assert "error" in response.json()

# Тест 7: Проверка health эндпоинта (агрегированный)
def test_health_endpoint():
    """Тест проверки здоровья обоих сервисов"""
    with patch('httpx.AsyncClient.get') as mock_get:
        mock_get.return_value = AsyncMock(
            status_code=200,
            json=AsyncMock(return_value={"status": "ok"})
        )
        
        response = client.get("/api/health")
        
        assert response.status_code == 200
        assert response.json()["go_service"] == "healthy"
        assert response.json()["fastapi"] == "healthy"

# Тест 8: Валидация входных данных в FastAPI
def test_validate_user_input():
    """Тест валидации входных данных для создания пользователя"""
    invalid_data = {"wrong_field": "test"}
    
    response = client.post("/api/users", json=invalid_data)
    
    assert response.status_code == 422  # Validation error

# Интеграционный тест (требует запущенный Go сервис)
@pytest.mark.integration
def test_integration_real_go_service():
    """Интеграционный тест с реальным Go сервисом"""
    # Этот тест требует запущенный Go сервис на localhost:8080
    response = client.get("/api/users/1")
    assert response.status_code in [200, 404]  # Может быть либо найден, либо нет
    
    health_response = client.get("/api/health")
    assert health_response.status_code == 200