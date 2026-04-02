import pytest
from fastapi.testclient import TestClient
from unittest.mock import AsyncMock, patch
import httpx
from main import app

client = TestClient(app)

@pytest.mark.asyncio
async def test_proxy_get_user_success_mock():
    """Тест успешного получения пользователя с моком"""
    mock_response = {"id": "123", "name": "User 123"}
    
    with patch('httpx.AsyncClient.get') as mock_get:
        mock_get.return_value = AsyncMock(
            status_code=200,
            json=AsyncMock(return_value=mock_response)
        )
        
        response = client.get("/api/users/123")
        
        assert response.status_code == 200
        assert response.json() == mock_response

@pytest.mark.asyncio
async def test_proxy_get_user_404_mock():
    """Тест обработки 404 с моком"""
    with patch('httpx.AsyncClient.get') as mock_get:
        mock_get.return_value = AsyncMock(
            status_code=404,
            json=AsyncMock(return_value={"error": "user not found"})
        )
        
        response = client.get("/api/users/999")
        
        assert response.status_code == 404
        assert "error" in response.json()

@pytest.mark.asyncio
async def test_proxy_create_user_success_mock():
    """Тест успешного создания пользователя с моком"""
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

@pytest.mark.asyncio
async def test_go_service_timeout_mock():
    """Тест обработки таймаута с моком"""
    with patch('httpx.AsyncClient.get') as mock_get:
        mock_get.side_effect = httpx.TimeoutException("Timeout")
        
        response = client.get("/api/users/123")
        
        assert response.status_code == 504
        assert "timeout" in response.json()["error"].lower()

@pytest.mark.asyncio
async def test_go_service_unavailable_mock():
    """Тест обработки недоступности с моком"""
    with patch('httpx.AsyncClient.get') as mock_get:
        mock_get.side_effect = httpx.ConnectError("Connection refused")
        
        response = client.get("/api/users/123")
        
        assert response.status_code == 503
        assert "unavailable" in response.json()["error"].lower()

@pytest.mark.asyncio
async def test_validate_user_input_mock():
    """Тест валидации входных данных"""
    invalid_data = {"wrong_field": "test"}
    
    response = client.post("/api/users", json=invalid_data)
    
    assert response.status_code == 422